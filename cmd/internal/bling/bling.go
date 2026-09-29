package bling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/bracomil/bracomil-internal-api-back/cmd/httpclient"
	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/logger"
	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/utils"
	"github.com/gofiber/fiber/v3"
)

func NewHandler(log *logger.Logger, client *httpclient.Client) *BlingHandler {
	// baseUrl := utils.GetEnv("BLING_API_URI", "https://api.bling.com.br/Api/v3")
	return &BlingHandler{
		mu:             sync.Mutex{},
		log:            log,
		client:         client,
		maxConcurrence: 3,
	}
}

func (h *BlingHandler) ListReceivables(c fiber.Ctx) error {
	page := c.Query("page", "1")
	limit := c.Query("limit", "100")

	filterType := c.Query("filterType", "V")
	date := c.Query("filterDate")

	return h.listReceivablesByDate(c, page, limit, filterType, date)
}

const (
	defaultPageSize = 100
	maxPages        = 100
)

// TODO: Remover responsabilidades do router da controller (uso do fiber.Ctx)
func (h *BlingHandler) listReceivablesByDate(c fiber.Ctx, pageStr, limitStr, filterType, dateStr string) error {
	// 1. Parse e validação
	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = defaultPageSize
	}

	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return err
	}

	query := url.Values{}
	query.Set("tipoFiltroData", "V")
	query.Set("dataInicial", date.Format("2006-01-02"))
	query.Set("dataFinal", date.Format("2006-01-02"))

	var result []AccountsReceivableDTO
	// 2. Paginação
	for p := page; p < page+maxPages; p++ {
		query.Set("page", strconv.Itoa(p))
		query.Set("limit", strconv.Itoa(limit))

		var pageResult Page[AccountsReceivableDTO]
		if err := h.client.DoJSON(c.Context(), http.MethodGet, "/contas/receber", query, nil, nil, &pageResult); err != nil {
			h.log.Error("failed to fetch receivables", "err", err, "page", p)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"message": "an error occurred while trying to get receivables",
				"error":   err.Error(),
			})
		}

		result = append(result, pageResult.Data...)

		// Última página: retornou menos que o limite
		if len(pageResult.Data) < limit {
			break
		}

	}

	// TODO: Apagar
	if utils.GetEnv("AMBIENT", "production") == "test" {
		substituteOriginValuesForTestValues(date.Format("2006-01-02"), result)
	}
	// 3. Resposta
	return c.JSON(fiber.Map{
		"data": result,
		"meta": fiber.Map{
			"total":  len(result),
			"cached": false,
		},
	})
}

func substituteOriginValuesForTestValues(date string, result []AccountsReceivableDTO) {
	switch date {
	case "2026-09-18":
		_ = []AccountsReceivableDTO{{ID: 26719286803, Status: 2, DueDate: "2026-09-18", Amount: 431.25, TransactionID: "", PixQRCodeLink: "", BoletoLink: "https://www.bling.com.br/doc.view.php?id=17a66ccd0c88aa9701cf4703540f2e157a038c183912c0b961355ac2427a9a17", IssueDate: "2026-08-28", Contact: AccountsReceivableContactDTO{ID: 15690732075, Name: "SYNPACK COMERCIO DE EMBALAGENS DO NORDESTE LTDA ME", Document: "19153331000190", PersonType: "J"}, PaymentMethod: &AccountsReceivablePaymentMethodDTO{ID: 5759352, FiscalCode: 15}, LedgerAccount: &AccountsReceivableLedgerAccountDTO{ID: 14890441138, Description: "BNB 300|025193-9 (COBRANÇA)"}, Origin: &AccountsReceivableOriginDTO{ID: 26719285562, OriginType: "notafiscal", Number: "019984", IssueDate: "2026-08-28", Amount: 862.5, Status: 7, URL: "https://api.bling.com.br/doc.view.php?id=76182770de8a1ebe11221f222cd4f978e7859d1bf1d324162326009b4a0890bd"}}}
		result[0].Origin = &AccountsReceivableOriginDTO{ID: 26719285562, OriginType: "notafiscal", Number: "019984", IssueDate: "2026-08-28", Amount: 862.5, Status: 7, URL: "https://api.bling.com.br/doc.view.php?id=76182770de8a1ebe11221f222cd4f978e7859d1bf1d324162326009b4a0890bd"}
	}
}

func (h *BlingHandler) SettleBatchReceipts(c fiber.Ctx) error {
	ctx := c.Context()
	var req SettleBatchRequest
	if err := json.Unmarshal(c.BodyRaw(), &req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request",
		})
	}
	if len(req.Items) < 1 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "empty request",
		})
	}

	result := &SettleBatchReceiptsResponse{}
	wg := sync.WaitGroup{}
	sem := make(chan struct{}, h.maxConcurrence)
	// Receber informações de carrier e category id
	for _, item := range req.Items {
		wg.Add(1)
		go func() {
			sem <- struct{}{}
			if err := h.getReceivableById(ctx, item); err != nil {
				result.Results = append(result.Results, SettleBatchReceiptsResponseItem{
					ReceivableID: item.Id,
					Status:       "error",
					Error: &BatchReceiptsResponse{
						Code:    "0",
						Message: fmt.Sprintf(err.Error()),
					},
				})
				result.Total++
				result.Failures++
			}
			wg.Done()
			<-sem
		}()
	}
	wg.Wait()

	for _, item := range req.Items {
		wg.Add(1)
		if item.Info.Carrier == nil || item.Info.Carrier.ID == 0 {
			continue
		}
		if item.Info.Category == nil || item.Info.Category.ID == 0 {
			continue
		}
		go func() {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			bordero, err := h.createReceiptForReceivable(ctx, item)
			if err != nil {
				result.Results = append(result.Results, SettleBatchReceiptsResponseItem{
					ReceivableID: item.Id,
					Status:       "error",
					Error: &BatchReceiptsResponse{
						Code:    "0",
						Message: fmt.Sprintf(err.Error()),
					},
				})
				result.Total++
				result.Failures++
				return
			}

			result.Results = append(result.Results, SettleBatchReceiptsResponseItem{
				ReceivableID: item.Id,
				Status:       "success",
				Bordero:      bordero,
			})
			result.Total++
			result.Successes++
		}()
	}
	wg.Wait()

	var status int
	switch {
	case result.Failures == 0:
		status = fiber.StatusOK
	case result.Failures > 0 && result.Successes > 0:
		status = fiber.StatusMultiStatus
	default:
		status = fiber.StatusInternalServerError
	}
	return c.Status(status).JSON(result)
}

func (h *BlingHandler) getReceivableById(ctx context.Context, item *SettleItemRequest) error {
	var nestedResponse DataResponse[AccountsReceivableDetailDTO]
	if err := h.client.DoJSON(ctx, http.MethodGet, fmt.Sprintf("/contas/receber/%d", item.Id), nil, nil, nil, &nestedResponse); err != nil {
		return fmt.Errorf("Could not get details from receivable. %w", err)
	}
	details := nestedResponse.Data

	if details.Category == nil || details.Category.ID == 0 {
		return errors.New("Invalid category in details of id")
	}

	if details.Carrier == nil || details.Carrier.ID == 0 {
		return errors.New("Invalid carrier in details of id")
	}
	// if details.Carrier.ID != CARRIER_ID_BNB_9 {
	// 	return errors.New("Receivable is registered in wrong carrier'")
	// }

	item.Info.Category = details.Category
	item.Info.Carrier = details.Carrier

	return nil
}

func (h *BlingHandler) createReceiptForReceivable(ctx context.Context, item *SettleItemRequest) (*AccountsReceivableBorderoDTO, error) {
	// Adjust values from centavos to reais
	item.Info.Addition = item.Info.Addition / 100
	item.Info.Discount = item.Info.Discount / 100
	item.Info.Fee = item.Info.Fee / 100
	item.Info.Interest = item.Info.Interest / 100
	item.Info.ReceivedValue = item.Info.ReceivedValue / 100

	// Add History
	item.Info.History = "Baixado automáticamente pela API"

	var response DataResponse[CreateReceiptForReceivableResponse]
	if err := h.client.DoJSON(ctx, fiber.MethodPost, fmt.Sprintf("/contas/receber/%d/baixar", item.Id), nil, nil, item.Info, &response); err != nil {
		return nil, err
	}
	return response.Data.Bordero, nil
}

func (h *BlingHandler) Temp(c fiber.Ctx) {
	zz := map[int]string{}
	wg := sync.WaitGroup{}
	for index := range 5 {
		wg.Add(1)
		go func() {
			var respModulos DataResponse[map[string]interface{}]
			if err := h.client.DoJSON(c.Context(), fiber.MethodGet, "/pedidos/vendas/26955389439", nil, nil, nil, &respModulos); err != nil {
				zz[index] = err.Error()
			}
			wg.Done()
		}()
	}
	wg.Wait()
	var status int
	switch {
	case len(zz) > 0:
		status = fiber.StatusMultiStatus
	case len(zz) == 0:
		status = fiber.StatusOK
	default:
		status = fiber.StatusBadGateway
	}
	c.Status(status).JSON(zz)
}
