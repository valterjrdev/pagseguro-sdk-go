package pagseguro

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/stretchr/testify/assert"
	"github.com/valterjrdev/pagseguro-sdk-go/pagseguro/models"
)

func TestPagseguro_Client(t *testing.T) {
	t.Run("create order", func(t *testing.T) {
		token := gofakeit.LetterN(60)
		order := &models.Order{
			ReferenceID: "ex-00001",
			Customer: models.Customer{
				Name:  "Jose da Silva",
				Email: "email@gmail.com",
				TaxID: "12345678909",
				Phones: []models.Phone{
					{
						Country: "55",
						Area:    "11",
						Number:  "999999999",
						Type:    "MOBILE",
					},
				},
			},
			Items: []models.Item{
				{
					ReferenceID: "referencia do item",
					Name:        "nome do item",
					Quantity:    1,
					UnitAmount:  500,
				},
			},
			Shipping: models.Shipping{
				Address: models.Address{
					Street:     "Avenida Brigadeiro Faria Lima",
					Number:     "1384",
					Locality:   "Pinheiros",
					City:       "São Paulo",
					Region:     "São Paulo",
					RegionCode: "SP",
					Country:    "BRA",
					PostalCode: "01452002",
				},
			},
			NotificationUrls: []string{"https://meusite.com/notificacoes"},
			Charges: []models.Charge{
				{
					ReferenceID: "referencia da cobranca",
					Description: "descricao da cobranca",
					Amount: models.Amount{
						Value:    500,
						Currency: "BRL",
					},
					PaymentMethod: models.PaymentMethod{
						Type: "BOLETO",
						Boleto: models.Boleto{
							DueDate: "2024-12-31",
							InstructionLines: models.InstructionLines{
								Line1: "Pagamento processado para DESC Fatura",
								Line2: "Via PagSeguro",
							},
							Holder: models.Holder{
								Name:  "Jose da Silva",
								TaxID: "22222222222",
								Email: "jose@email.com",
								Address: models.Address{
									Country:    "Brasil",
									Region:     "São Paulo",
									RegionCode: "SP",
									City:       "Sao Paulo",
									PostalCode: "01452002",
									Street:     "Avenida Brigadeiro Faria Lima",
									Number:     "1384",
									Locality:   "Pinheiros",
								},
							},
						},
					},
				},
			},
		}

		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, token, r.Header.Get("Authorization"))

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			w.Write([]byte(``))
		}))

		defer srv.Close()

		client := New(srv.URL, token)
		err := client.CreateOrder(context.Background(), order)
		assert.EqualError(t, err, "")
	})

	t.Run("error responses", func(t *testing.T) {
		t.Run("order standard", func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`
					{
						"code": "40001",
						"description": "required_parameter",
						"parameter_name": "payment_methods_is_required"
					}
				`))
			}))

			defer srv.Close()

			client := New(srv.URL, "")
			err := client.CreateOrder(context.Background(), &models.Order{})
			assert.EqualError(t, err, "error processing request(http status code: 400)")
			assert.Equal(
				t,
				[]ApiError{
					{Code: "40001", Description: "required_parameter", ParameterName: "payment_methods_is_required"},
				},
				err.(*ApiErrors).ErrorMessages,
			)
		})

		t.Run("charge standard", func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`
					{
						"error_messages": [
							{
								"code": "40001",
								"description": "required_parameter",
								"parameter_name": "payment_method.capture"
							},
							{
								"code": "40002",
								"description": "invalid_parameter",
								"parameter_name": "payment_methods_is_invalid"
							}
						]
					}
				`))
			}))

			defer srv.Close()

			client := New(srv.URL, "")
			err := client.CreateOrder(context.Background(), &models.Order{})
			assert.EqualError(t, err, "error processing request(http status code: 400)")
			assert.Equal(
				t,
				[]ApiError{
					{Code: "40001", Description: "required_parameter", ParameterName: "payment_method.capture"},
					{Code: "40002", Description: "invalid_parameter", ParameterName: "payment_methods_is_invalid"},
				},
				err.(*ApiErrors).ErrorMessages,
			)
		})

		t.Run("internal server error", func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte(``))
			}))

			defer srv.Close()

			client := New(srv.URL, "")
			err := client.CreateOrder(context.Background(), &models.Order{})
			assert.EqualError(t, err, "error processing request(http status code: 500): non-standard error response, contact pagseguro support")
		})

		t.Run("non-standard error response", func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`
					{
						"teste_message": "test"
					}
				`))
			}))

			defer srv.Close()

			client := New(srv.URL, "")
			err := client.CreateOrder(context.Background(), &models.Order{})
			assert.EqualError(t, err, "error processing request(http status code: 400): non-standard error response, contact pagseguro support")
		})
	})
}