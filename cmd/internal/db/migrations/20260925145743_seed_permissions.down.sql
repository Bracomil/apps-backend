DELETE FROM permissions WHERE name IN (
    'bank-returns:read',
    'bank-returns:settle',
    'ofx:read'
);