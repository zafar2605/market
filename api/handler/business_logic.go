package handler

import (
	"context"
	"errors"
	"fmt"
	"market_system/config"
	"market_system/models"
	"market_system/pkg/helpers"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func (h *Handler) SaleScanBarcode(c *gin.Context) {

	var (
		saleID   = c.Query("sale_id")
		branchID = c.Query("branch_id")
		barcode  = c.Query("barcode")
	)
	if !helpers.IsValidUUID(saleID) {
		handleResponse(c, http.StatusBadRequest, "sale id is not uuid")
		return
	}

	if !helpers.IsValidUUID(branchID) {
		handleResponse(c, http.StatusBadRequest, "branch id is not uuid")
		return
	}

	remainingTableProduct, err := h.strg.Remainder().GetList(context.Background(), &models.GetListRemainderRequest{
		Limit: 1,
		Query: fmt.Sprintf(" AND bracode = %s AND branch_id = %s", barcode, branchID),
	})

	if err != nil {
		handleResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	if len(remainingTableProduct.Remainder) <= 0 {
		handleResponse(c, http.StatusBadRequest, "Товар не найден")
		return
	}

	saleProduct, err := h.strg.Sale_Product().GetList(context.Background(), &models.GetListSaleProductRequest{
		Limit: 1,
		Query: fmt.Sprintf(" AND bracode = %s AND sale_id = %s", barcode, saleID),
	})

	if err != nil {
		handleResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	if len(saleProduct.SaleProducts) <= 0 {
		var product = remainingTableProduct.Remainder[0]
		_, err := h.strg.Sale_Product().Create(context.Background(), &models.CreateSaleProduct{
			SaleID:            saleID,
			CategoryID:        product.CategoryID,
			ProductName:       product.ProductName,
			Barcode:           product.Barcode,
			RemainingQuantity: product.Quantity,
			Quantity:          1,
			AllowDiscount:     false,
			DiscountType:      "",
			Discount:          0,
			Price:             product.PriceIncome,
			TotalAmount:       product.PriceIncome,
		})
		if err != nil {
			handleResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		handleResponse(c, http.StatusCreated, "Успешно")
		return
	}

	if remainingTableProduct.Remainder[0].Quantity-saleProduct.SaleProducts[0].Quantity < 0 {
		handleResponse(c, http.StatusBadRequest, "Максималный лимит")
		return
	}

	_, err = h.strg.Sale_Product().Update(context.Background(), &models.UpdateSaleProduct{
		Id:                saleProduct.SaleProducts[0].Id,
		RemainingQuantity: saleProduct.SaleProducts[0].RemainingQuantity,
		Quantity:          saleProduct.SaleProducts[0].Quantity + 1,
		AllowDiscount:     saleProduct.SaleProducts[0].AllowDiscount,
		DiscountType:      saleProduct.SaleProducts[0].DiscountType,
		Discount:          saleProduct.SaleProducts[0].Discount,
		Price:             saleProduct.SaleProducts[0].Price,
		TotalAmount:       saleProduct.SaleProducts[0].TotalAmount + saleProduct.SaleProducts[0].Price,
	})

	if err != nil {
		handleResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	handleResponse(c, http.StatusCreated, "Успешно")
	return
}

func (h *Handler) Dosale(c *gin.Context) {
	var (
		saleID   = c.Query("sale_id")
		branchID = c.Query("branch_id")
	)
	if !helpers.IsValidUUID(saleID) {
		handleResponse(c, http.StatusBadRequest, "sale id is not uuid")
		return
	}

	if !helpers.IsValidUUID(branchID) {
		handleResponse(c, http.StatusBadRequest, "branch id is not uuid")
		return
	}

	saleData, err := h.strg.Sale().GetByID(context.Background(), &models.SalePrimaryKey{Id: saleID})
	if err != nil {
		handleResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	salePaymentResponse, err := h.strg.Payment().GetList(context.Background(), &models.GetListPaymentRequest{
		Limit: 100,
		Query: fmt.Sprintf(" AND sale_id = %s", saleID),
	})
	if err != nil {
		handleResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	cashTransactionResponse, err := h.strg.Transaction().GetList(context.Background(), &models.GetListTransactonRequest{
		Limit: 100,
		Query: fmt.Sprintf(" AND shift_id = %s", saleData.ShiftID),
	})
	if err != nil {
		handleResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	if len(cashTransactionResponse.Transactions) <= 0 {
		handleResponse(c, http.StatusBadRequest, "не найден транзакции")
		return
	}

	var (
		salePayment     = salePaymentResponse.Payments[0]
		cashTransaction = cashTransactionResponse.Transactions[0]
	)
	_, err = h.strg.Transaction().Update(context.Background(), &models.UpdateTransaction{
		Id:          cashTransaction.Id,
		Cash:        cashTransaction.Cash + salePayment.Cash,
		Uzcard:      cashTransaction.Uzcard + salePayment.Uzcard,
		Payme:       cashTransaction.Payme + salePayment.Payme,
		Click:       cashTransaction.Click + salePayment.Click,
		Humo:        cashTransaction.Humo + salePayment.Humo,
		Apelsin:     cashTransaction.Apelsin + salePayment.Apelsin,
		TotalAmount: cashTransaction.TotalAmount + salePayment.TotalAmount,
	})
	if err != nil {
		handleResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	saleProductResponse, err := h.strg.Sale_Product().GetList(context.Background(), &models.GetListSaleProductRequest{
		Limit: 1000,
		Query: fmt.Sprintf(" AND sale_id = %s", saleID),
	})
	if err != nil {
		handleResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	var remainderQuery = fmt.Sprintf(" AND branch_id = %s AND barcode IN (", branchID)
	for _, saleProduct := range saleProductResponse.SaleProducts {
		remainderQuery += fmt.Sprintf("'%s',", saleProduct.Barcode)
	}
	remainderQuery = remainderQuery[:len(remainderQuery)-1]
	remainderQuery += ")"

	remainderResponse, err := h.strg.Remainder().GetList(context.Background(), &models.GetListRemainderRequest{
		Limit: 1000,
		Query: remainderQuery,
	})
	if err != nil {
		handleResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	for _, saleProduct := range saleProductResponse.SaleProducts {
		for _, remainder := range remainderResponse.Remainder {
			if saleProduct.Barcode == remainder.Barcode {
				_, err := h.strg.Remainder().Update(context.Background(), &models.UpdateRemainder{
					Id:          remainder.Id,
					ProductName: remainder.ProductName,
					Barcode:     remainder.Barcode,
					PriceIncome: remainder.PriceIncome,
					Quantity:    remainder.Quantity - saleProduct.Quantity,
				})
				if err != nil {
					handleResponse(c, http.StatusInternalServerError, err.Error())
					return
				}
			}
		}
	}

	_, err = h.strg.Sale().Update(context.Background(), &models.UpdateSale{
		Id:          saleData.Id,
		BranchID:    saleData.BranchID,
		SalePointID: saleData.SalePointID,
		ShiftID:     saleData.ShiftID,
		EmployeeID:  saleData.EmployeeID,
		Barcode:     saleData.Barcode,
		Status:      "finished",
	})
	if err != nil {
		handleResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	handleResponse(c, http.StatusCreated, "Успешно")
	return
}

func (h *Handler) DoIncome(c *gin.Context) {

	var (
		coming_id = c.Query("coming_id")
	)
	ctx, cencel := context.WithTimeout(context.Background(), config.CtxTimeout)
	defer cencel()

	incomeTable, err := h.strg.Income().GetByID(ctx, &models.IncomePrimaryKey{Id: coming_id})
	if err != nil {
		handleResponse(c, http.StatusBadRequest, err.Error())
		return
	}
	if incomeTable.Status == "finished" {
		handleResponse(c, http.StatusBadRequest, errors.New("coming table status finished"))
		return
	}
	remainderTable, err := h.strg.Remainder().GetByID(ctx, &models.RemainderPrimaryKey{Id: incomeTable.BranchID})
	if err != nil {
		handleResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	incriseRemainder, err := h.strg.Remainder().GetList(ctx, &models.GetListRemainderRequest{
		Query: fmt.Sprintf(" AND branch_id = %s AND product_name = %s", incomeTable.BranchID, remainderTable.ProductName),
	})
	if err != nil {
		handleResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	var (
		remainder      = models.Remainder{}
		income_product = models.IncomeProduct{}
	)
	remainder = *incriseRemainder.Remainder[0]
	// for _, v := range incriseRemainder.Remainder {
	// 	remainder = models.Remainder{
	// 		Id:          v.Id,
	// 		BranchID:    v.BranchID,
	// 		CategoryID:  v.CategoryID,
	// 		ProductName: v.ProductName,
	// 		Barcode:     v.Barcode,
	// 		PriceIncome: v.PriceIncome,
	// 		Quantity:    v.Quantity,
	// 		CreatedAt:   v.CreatedAt,
	// 		UpdatedAt:   v.UpdatedAt,
	// 	}
	// }

	incomeProductList, err := h.strg.IncomeProduct().GetList(ctx, &models.GetListIncomeProductRequest{
		Limit: 1,
		Query: fmt.Sprintf(" AND income_id = %s", incomeTable.Id),
	})
	if err != nil {
		handleResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	income_product = *incomeProductList.IncomeProducts[0]
	remainder = models.Remainder{
		Id:          remainder.Id,
		BranchID:    remainder.BranchID,
		CategoryID:  remainder.CategoryID,
		ProductName: remainder.ProductName,
		Barcode:     remainder.Barcode,
		PriceIncome: remainder.PriceIncome,
		Quantity:    remainder.Quantity + int(income_product.Quantity),
	}

	_, err = h.strg.Remainder().Update(ctx, &models.UpdateRemainder{
		Id:          remainder.Id,
		ProductName: remainder.ProductName,
		PriceIncome: remainder.PriceIncome,
		Barcode:     remainder.Barcode,
		Quantity:    remainder.Quantity,
	})
	if err != nil {
		handleResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

}

func (h *Handler) ShiftTable(c *gin.Context) {

	var (
		method      = c.Query("method")
		cashTableId = c.Query("cash_table")
	)

	ctx, cencel := context.WithTimeout(context.Background(), config.CtxTimeout)
	defer cencel()

	shiftResp, err := h.strg.Shift().GetByID(ctx, &models.ShiftPrimaryKey{
		Id: cashTableId,
	})
	if err != nil {
		handleResponse(c, http.StatusBadRequest, err.Error())
		return
	}

	if method == "open" {

		_, err = h.strg.Transaction().Create(ctx, &models.CreateTransaction{
			ShiftID:     shiftResp.Id,
			Cash:        0,
			Uzcard:      0,
			Payme:       0,
			Click:       0,
			Humo:        0,
			Apelsin:     0,
			TotalAmount: 0,
		})
		if err != nil {
			handleResponse(c, http.StatusInternalServerError, err.Error())
			return
		}
		_, err = h.strg.Shift().Update(ctx, &models.UpdateShift{
			Id:          cashTableId,
			BranchID:    shiftResp.BranchID,
			UserID:      shiftResp.UserID,
			SalePointID: shiftResp.SalePointID,
			Status:      "Открытая",
			OpenShift:   time.Now().Format("2006-01-02 15:04:05"),
			CloseShift:  shiftResp.CloseShift,
		})
		if err != nil {
			handleResponse(c, http.StatusBadRequest, err.Error())
			return
		}
	} else if method == "close" {

		_, err = h.strg.Shift().Update(ctx, &models.UpdateShift{
			Id:          cashTableId,
			BranchID:    shiftResp.BranchID,
			UserID:      shiftResp.UserID,
			SalePointID: shiftResp.SalePointID,
			Status:      "Закрытая",
			OpenShift:   shiftResp.OpenShift,
			CloseShift:  time.Now().Format("2006-01-02 15:04:05"),
		})
		if err != nil {
			handleResponse(c, http.StatusBadRequest, err.Error())
			return
		}
	}

}
