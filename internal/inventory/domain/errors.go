package domain

import "errors"

var (
	ErrWarehouseNotFound       = errors.New("warehouse not found")
	ErrBinNotFound             = errors.New("bin location not found")
	ErrDuplicateWarehouseCode  = errors.New("duplicate warehouse code for legal entity")
	ErrDuplicateBinCode        = errors.New("duplicate bin location code for warehouse")
	ErrInsufficientStock       = errors.New("insufficient available stock")
	ErrNegativeStockNotAllowed = errors.New("negative stock is not allowed by policy")
	ErrInvalidMovementType     = errors.New("invalid stock movement type")
	ErrDuplicateMovementNumber = errors.New("duplicate stock movement number")
	ErrMovementImmutable       = errors.New("posted stock movements are immutable")
	ErrReservationNotFound     = errors.New("stock reservation not found or not active")
	ErrTransferOrderNotFound   = errors.New("transfer order not found")
	ErrInvalidTransferState    = errors.New("invalid transfer order state transition")
	ErrStockAdjustmentNotFound = errors.New("stock adjustment not found")
	ErrStockCountNotFound      = errors.New("stock count not found")
	ErrLotNotFound             = errors.New("lot number not found")
	ErrSerialNumberNotFound    = errors.New("serial number not found")
)
