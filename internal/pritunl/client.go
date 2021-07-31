package pritunl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type Client interface {
	GetOrganizationByID(id string) (*Organization, error)
	GetOrganization(name string) (*Organization, error)
	CreateOrganization(name string) (*Organization, error)
	RenameOrganization(id string, name string) error
	DeleteOrganization(name string) error

	GetServer(id string) (*Server, error)
	CreateServer(name, protocol, cipher, hash string, port *int) (*Server, error)
	UpdateServer(id string, updatedServer *Server) error
	DeleteServer(id string) error
	AttachOrganizationToServer(organizationId, serverId string) error
	DetachOrganizationFromServer(organizationId, serverId string) error

	StartServer(serverId string) error
	StopServer(serverId string) error
	//RestartServer(serverId string) error
	//DeleteServer(serverId string) error

	AddRouteToServer(serverId string, route Route) error
	DeleteRouteFromServer(serverId string, route Route) error
	//UpdateRouteOnServer(serverId string, route Route) error
}

type client struct {
	httpClient *http.Client
	baseUrl    string
}

func (c client) GetOrganizationByID(id string) (*Organization, error) {
	url := fmt.Sprintf("/organization/%s", id)
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	// iterate over all pages
	var organization Organization

	err = json.Unmarshal(body, &organization)
	if err != nil {
		return nil, fmt.Errorf("GetOrganizationByID: %s: %+v, id=%s, body=%s", err, organization, id, body)
	}

	return &organization, nil
}

func (c client) GetOrganization(name string) (*Organization, error) {
	url := "/organization"
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	type GetOrganizationsApiResponse struct {
		Organizations []Organization
	}

	var organizations []Organization

	err = json.Unmarshal(body, &organizations)
	if err != nil {
		return nil, fmt.Errorf("GetOrganization: %s: %+v, name=%s, body=%s", err, organizations, name, body)
	}

	for _, organization := range organizations {
		if strings.ToLower(organization.Name) == strings.ToLower(name) {
			return &organization, nil
		}
	}

	return nil, nil
}

func (c client) CreateOrganization(name string) (*Organization, error) {
	var jsonStr = []byte(`{"name": "` + name + `"}`)

	url := "/organization"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	var organization Organization
	err = json.Unmarshal(body, &organization)
	if err != nil {
		return nil, fmt.Errorf("CreateOrganization: %s: %+v, name=%s, body=%s", err, organization, name, body)
	}

	return &organization, nil
}

func (c client) RenameOrganization(id string, name string) error {
	panic("implement me")
}

func (c client) DeleteOrganization(id string) error {
	url := fmt.Sprintf("/organization/%s", id)
	req, err := http.NewRequest("DELETE", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	var organization Organization
	err = json.Unmarshal(body, &organization)
	if err != nil {
		return fmt.Errorf("DeleteOrganization: %s: %+v, id=%s, body=%s", err, organization, id, body)
	}

	return nil
}

func (c client) GetServer(id string) (*Server, error) {
	url := fmt.Sprintf("/server/%s", id)
	req, err := http.NewRequest("GET", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GetServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var server Server
	err = json.Unmarshal(body, &server)

	if err != nil {
		return nil, fmt.Errorf("GetServer: %s: %+v, id=%s, body=%s", err, server, id, body)
	}

	return &server, nil
}

func (c client) CreateServer(name, protocol, cipher, hash string, port *int) (*Server, error) {
	serverStruct := Server{
		Name:     name,
		Protocol: protocol,
		Cipher:   cipher,
		Hash:     hash,
	}

	if port != nil {
		serverStruct.Port = *port
	}

	jsonData, err := json.Marshal(serverStruct)

	url := "/server"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("CreateServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	var server Server
	err = json.Unmarshal(body, &server)
	if err != nil {
		return nil, err
	}

	return &server, nil
}

func (c client) UpdateServer(id string, updatedServer *Server) error {
	jsonData, err := json.Marshal(updatedServer)
	if err != nil {
		return fmt.Errorf("UpdateServer: Error on marshalling data: %s [data=%+v]", err, updatedServer)
	}

	url := fmt.Sprintf("/server/%s", id)
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("UpdateServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	return nil
}

func (c client) DeleteServer(id string) error {
	url := fmt.Sprintf("/server/%s", id)
	req, err := http.NewRequest("DELETE", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("DeleteServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	var server Server
	err = json.Unmarshal(body, &server)
	if err != nil {
		return fmt.Errorf("DeleteServer: Error on parsing response: %s (id=%s, body=%s)", err, id, body)
	}

	return nil
}

func (c client) AttachOrganizationToServer(organizationId, serverId string) error {
	url := fmt.Sprintf("/server/%s/organization/%s", serverId, organizationId)
	req, err := http.NewRequest("PUT", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	return nil
}

func (c client) DetachOrganizationFromServer(organizationId, serverId string) error {
	url := fmt.Sprintf("/server/%s/organization/%s", serverId, organizationId)
	req, err := http.NewRequest("DELETE", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	return nil
}

func (c client) StartServer(serverId string) error {
	url := fmt.Sprintf("/server/%s/operation/start", serverId)
	req, err := http.NewRequest("PUT", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	return nil
}

func (c client) StopServer(serverId string) error {
	url := fmt.Sprintf("/server/%s/operation/stop", serverId)
	req, err := http.NewRequest("PUT", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	return nil
}

func (c client) AddRouteToServer(serverId string, route Route) error {
	err := c.StopServer(serverId)
	if err != nil {
		return err
	}

	jsonData, err := json.Marshal(route)

	url := fmt.Sprintf("/server/%s/route", serverId)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("AddRouteToServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	err = c.StartServer(serverId)
	if err != nil {
		return err
	}

	return nil
}

func (c client) DeleteRouteFromServer(serverId string, route Route) error {
	err := c.StopServer(serverId)
	if err != nil {
		return fmt.Errorf("DeleteRouteFromServer: Error on stopping the server reqeust: %s", err)
	}

	url := fmt.Sprintf("/server/%s/route/%s", serverId, route.ID)
	req, err := http.NewRequest("DELETE", url, nil)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("DeleteRouteFromServer: Error on HTTP request: %s", err)
	}
	defer resp.Body.Close()

	err = c.StartServer(serverId)
	if err != nil {
		return fmt.Errorf("DeleteRouteFromServer: Error on starting the server reqeust: %s", err)
	}

	return nil
}

func NewClient(baseUrl, apiToken, apiSecret string) Client {
	httpClient := &http.Client{
		Transport: &transport{
			baseUrl:             baseUrl,
			apiToken:            apiToken,
			apiSecret:           apiSecret,
			underlyingTransport: http.DefaultTransport,
		},
	}

	return &client{httpClient: httpClient}
}
