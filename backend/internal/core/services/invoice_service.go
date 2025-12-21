package services

import (
	"kubecloud/internal/billing"
	cfg "kubecloud/internal/config"
	"kubecloud/internal/core/models"
)

type InvoiceService struct {
	invoicesRepo       models.InvoiceRepository
	userRepo           models.UserRepository
	invoiceCompanyData cfg.InvoiceCompanyData
}

func NewInvoiceService(
	invoiceRepo models.InvoiceRepository, userRepo models.UserRepository,
	invoiceCompanyData cfg.InvoiceCompanyData,
) InvoiceService {
	return InvoiceService{
		invoicesRepo:       invoiceRepo,
		userRepo:           userRepo,
		invoiceCompanyData: invoiceCompanyData,
	}
}

func (svc *InvoiceService) ListInvoices() ([]models.Invoice, error) {
	return svc.invoicesRepo.ListInvoices()
}

func (svc *InvoiceService) ListUserInvoices(userID int) ([]models.Invoice, error) {
	return svc.invoicesRepo.ListUserInvoices(userID)
}

func (svc *InvoiceService) GetInvoiceByID(invoiceID, userID int) (models.Invoice, error) {
	invoice, err := svc.invoicesRepo.GetInvoice(invoiceID)
	if err != nil {
		return models.Invoice{}, err
	}

	// Creating pdf for invoice if it doesn't have it
	if len(invoice.FileData) == 0 {
		user, err := svc.userRepo.GetUserByID(userID)
		if err != nil {
			return models.Invoice{}, err
		}

		pdfContent, err := billing.CreateInvoicePDF(invoice, user, svc.invoiceCompanyData)
		if err != nil {
			return models.Invoice{}, err
		}

		invoice.FileData = pdfContent
		if err := svc.invoicesRepo.UpdateInvoicePDF(invoiceID, invoice.FileData); err != nil {
			return models.Invoice{}, err
		}
	}

	return invoice, nil
}
