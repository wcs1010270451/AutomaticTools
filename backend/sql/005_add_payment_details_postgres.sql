ALTER TABLE orders
    ADD COLUMN IF NOT EXISTS provider_trade_no VARCHAR(64) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payment_qr_code TEXT NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS payment_expires_at BIGINT;

COMMENT ON COLUMN orders.provider_trade_no IS '支付渠道交易号，例如支付宝交易号';
COMMENT ON COLUMN orders.payment_qr_code IS '支付渠道返回的二维码内容，仅待支付阶段使用';
COMMENT ON COLUMN orders.payment_expires_at IS '支付二维码失效时间，Unix秒级时间戳';

CREATE UNIQUE INDEX IF NOT EXISTS uk_orders_provider_trade_no
    ON orders(provider_trade_no)
    WHERE provider_trade_no <> '';
