package client

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/loan-service/internal/model"
)

type ServiceClient struct {
	userServiceURL string
	bookServiceURL string
	httpClient     *http.Client
}

func New(userServiceURL, bookServiceURL string) *ServiceClient {
	return &ServiceClient{
		userServiceURL: userServiceURL,
		bookServiceURL: bookServiceURL,
		httpClient:     &http.Client{},
	}
}

func (c *ServiceClient) GetUser(id int) (*model.User, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/users/%d", c.userServiceURL, id))
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar com User Service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("usuário não encontrado (status %d)", resp.StatusCode)
	}

	var user model.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta do User Service: %v", err)
	}
	return &user, nil
}

func (c *ServiceClient) GetBook(id int) (*model.Book, error) {
	resp, err := c.httpClient.Get(fmt.Sprintf("%s/api/books/%d", c.bookServiceURL, id))
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar com Book Service: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("livro não encontrado (status %d)", resp.StatusCode)
	}

	var book model.Book
	if err := json.NewDecoder(resp.Body).Decode(&book); err != nil {
		return nil, fmt.Errorf("erro ao decodificar resposta do Book Service: %v", err)
	}
	return &book, nil
}
