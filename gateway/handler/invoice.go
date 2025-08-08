package handler

import (
	"net/http"
	"strconv"

	"github.com/shamssahal/toll-calculator/aggregator/client"
	"github.com/shamssahal/toll-calculator/gateway/utils"
)

type InvoiceHandler struct {
	client *client.HTTPClient
}

func NewInvoiceHandler(c *client.HTTPClient) *InvoiceHandler {
	return &InvoiceHandler{
		client: c,
	}
}

func (h *InvoiceHandler) HandleGetInvoice(w http.ResponseWriter, r *http.Request) error {
	id := r.URL.Query().Get("id")
	_, err := strconv.Atoi(id)
	if err != nil {
		return utils.WriteJSON(w, http.StatusBadRequest,
			map[string]string{"error": "missing or incorrect 'id' query parameter"})
	}
	invoiceData, err := h.client.Invoice(r.Context(), id)
	if err != nil {
		return utils.WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to fetch invoice data"})
	}
	return utils.WriteJSON(w, http.StatusOK, invoiceData)
}
