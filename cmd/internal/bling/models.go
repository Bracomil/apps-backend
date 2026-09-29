package bling

import (
	"sync"

	"github.com/bracomil/bracomil-internal-api-back/cmd/httpclient"
	"github.com/bracomil/bracomil-internal-api-back/cmd/pkg/logger"
)

const CARRIER_ID_BNB_9 = 14890441138
const CARRIER_ID_BNB_9_TEST = 9

type BlingHandler struct {
	mu             sync.Mutex
	log            *logger.Logger
	client         *httpclient.Client
	maxConcurrence int32
}

type DataResponse[T any] struct {
	Data T `json:"data"`
}

type Page[T any] struct {
	Data  []T `json:"data"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type ReceivableQuery struct {
	Page            int
	Limit           int
	Situations      []int
	DateFilterType  string
	StartDate       string
	EndDate         string
	CategoryIDs     []int64
	CarrierID       int64
	ContactID       int64
	SellerID        int64
	PaymentMethodID int64
}

const (
	ReceivableStatusOpen              = 1 // Em aberto
	ReceivableStatusReceived          = 2 // Recebido
	ReceivableStatusPartiallyReceived = 3 // Parcialmente recebido
	ReceivableStatusReturned          = 4 // Devolvido
	ReceivableStatusPartiallyReturned = 5 // Parcialmente devolvido
	ReceivableStatusCancelled         = 6 // Cancelado
	ReceivableStatusConfirmed         = 7 // Confirmado
)

// AccountsReceivableDTO represents an account receivable in the listing
type AccountsReceivableDTO struct {
	ID            int64                               `json:"id,omitempty"`
	Status        int                                 `json:"situacao"`
	DueDate       string                              `json:"vencimento"`
	Amount        float64                             `json:"valor"`
	TransactionID string                              `json:"idTransacao,omitempty"`
	PixQRCodeLink string                              `json:"linkQRCodePix,omitempty"`
	BoletoLink    string                              `json:"linkBoleto,omitempty"`
	IssueDate     string                              `json:"dataEmissao,omitempty"`
	Contact       AccountsReceivableContactDTO        `json:"contato"`
	PaymentMethod *AccountsReceivablePaymentMethodDTO `json:"formaPagamento,omitempty"`
	LedgerAccount *AccountsReceivableLedgerAccountDTO `json:"contaContabil,omitempty"`
	Origin        *AccountsReceivableOriginDTO        `json:"origem,omitempty"`
}

// AccountsReceivableContactDTO represents the contact linked to the account
type AccountsReceivableContactDTO struct {
	ID         int64  `json:"id"`
	Name       string `json:"nome,omitempty"`
	Document   string `json:"numeroDocumento,omitempty"`
	PersonType string `json:"tipo,omitempty"` // "F" (individual) or "J" (company)
}

// AccountsReceivablePaymentMethodDTO represents the payment method
type AccountsReceivablePaymentMethodDTO struct {
	ID         int64 `json:"id"`
	FiscalCode int   `json:"codigoFiscal,omitempty"`
}

// AccountsReceivableLedgerAccountDTO represents the ledger account
type AccountsReceivableLedgerAccountDTO struct {
	ID          int64  `json:"id"`
	Description string `json:"descricao,omitempty"`
}

// AccountsReceivableOriginDTO represents the origin of the account (sale or invoice)
type AccountsReceivableOriginDTO struct {
	ID         int64   `json:"id"`
	OriginType string  `json:"tipoOrigem,omitempty"` // "venda" or "notaFiscal"
	Number     string  `json:"numero,omitempty"`
	IssueDate  string  `json:"dataEmissao,omitempty"`
	Amount     float64 `json:"valor,omitempty"`
	Status     int     `json:"situacao,omitempty"`
	URL        string  `json:"url,omitempty"`
}

type AccountsReceivableDetailDTO struct {
	AccountsReceivableDTO // embed: id, situacao, vencimento, valor, contato, etc.

	// ----- Campos do ContasReceberDadosBaseDTO -----
	Balance         float64                           `json:"saldo,omitempty"`
	OriginalDueDate string                            `json:"vencimentoOriginal,omitempty"`
	DocumentNumber  string                            `json:"numeroDocumento,omitempty"`
	Competence      string                            `json:"competencia,omitempty"`
	History         string                            `json:"historico,omitempty"`
	BankNumber      string                            `json:"numeroBanco,omitempty"`
	Carrier         *AccountsReceivableCarrierDTO     `json:"portador,omitempty"`
	Category        *AccountsReceivableCategoryDTO    `json:"categoria,omitempty"`
	Salesperson     *AccountsReceivableSalespersonDTO `json:"vendedor,omitempty"`
	Borderos        []int64                           `json:"borderos,omitempty"`

	// ----- Campos do ContasReceberDadosDTO -----
	Recurrence *AccountsReceivableRecurrenceDTO `json:"ocorrencia,omitempty"`
}

// AccountsReceivableCarrierDTO — portador
type AccountsReceivableCarrierDTO struct {
	ID int64 `json:"id"`
}

// AccountsReceivableCategoryDTO — categoria de receita/despesa
type AccountsReceivableCategoryDTO struct {
	ID int64 `json:"id"`
}

// AccountsReceivableSalespersonDTO — vendedor
type AccountsReceivableSalespersonDTO struct {
	ID int64 `json:"id"`
}

type AccountsReceivableRecurrenceDTO struct {
	Type int `json:"tipo"`

	// Parcelada (tipo=2)
	ConsiderWorkingDays bool `json:"considerarDiasUteis,omitempty"`
	DueDay              int  `json:"diaVencimento,omitempty"`
	Installments        int  `json:"numeroParcelas,omitempty"`

	// Mensal/Bimestral/etc (tipo=3..8)
	LimitDate string `json:"dataLimite,omitempty"`

	// Semanal (tipo=9)
	WeekdayDueDay int `json:"diaSemanaVencimento,omitempty"`
}

type AccountsReceivableBorderoDTO struct {
	ID int64 `json:"id"`
}

// ----------------------------------------------------  REQUESTS ---------------------------------------------------- //

type SettleBatchRequest struct {
	Items []*SettleItemRequest `json:"baixas" validate:"required,min=1,max=100,dive"`
}

type SettleItemRequest struct {
	Id   int64                         `json:"id"`
	Info BlingCreateReceiptRequestBody `json:"info"`
}

type BlingCreateReceiptRequestBody struct {
	Date          string                         `json:"data" validate:"required,datetime=2006-01-02"`
	ReceivedValue float64                        `json:"valorRecebido" validate:"required,gt=0"`
	Carrier       *AccountsReceivableCarrierDTO  `json:"portador"`  // Valor adquirido em buscas ao BLING
	Category      *AccountsReceivableCategoryDTO `json:"categoria"` // Valor adquirido em buscas ao BLING
	History       string                         `json:"historico"`
	Interest      float64                        `json:"juros"`
	Discount      float64                        `json:"desconto"`
	Addition      float64                        `json:"acrescimo"`
	Fee           float64                        `json:"tarifa"`
}

// ----------------------------------------------------  RESPONSE ---------------------------------------------------- //

type SettleBatchReceiptsResponse struct {
	Total     int32                             `json:"total"`
	Successes int32                             `json:"successes"`
	Failures  int32                             `json:"failures"`
	Results   []SettleBatchReceiptsResponseItem `json:"results"`
}

type SettleBatchReceiptsResponseItem struct {
	ReceivableID int64                         `json:"receivableId"`
	Status       string                        `json:"status"` // "success" | "error"
	Bordero      *AccountsReceivableBorderoDTO `json:"bordero,omitempty"`
	Error        *BatchReceiptsResponse        `json:"error,omitempty"`
}

type BatchReceiptsResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type CreateReceiptForReceivableResponse struct {
	Bordero *AccountsReceivableBorderoDTO `json:"bordero"`
}
