package solr

import (
	"encoding/json"
	"io"

	"github.com/pkg/errors"

	"fmt"
)

func (sc *SLClient) DecodeResponse(response *Response) (map[string]interface{}, error) {
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			err1 := errors.Wrap(err, "failed to parse response body")
			if err1 != nil {
				return
			}
			return
		}
	}(response.body)

	responseBody := make(map[string]interface{})
	if err := json.NewDecoder(response.body).Decode(&responseBody); err != nil {
		return nil, fmt.Errorf("failed to deserialize the response: %v", err)
	}

	return responseBody, nil
}

func (sc *SLClient) GetResponseStatus(responseBody map[string]interface{}) (int, error) {
	err, ok := responseBody["error"].(map[string]interface{})
	if ok {
		msg, ok := err["msg"].(string)
		if !ok {
			return -1, errors.New("no msg found in error message while getting response status")

		}
		code, ok := err["code"].(float64)
		if !ok {
			return -1, errors.New("error occurred but didn't found error code while getting response status")

		}
		return -1, errors.New(fmt.Sprintf("Error: %v with code %d", msg, int(code)))
	}
	responseHeader, ok := responseBody["responseHeader"].(map[string]interface{})
	if !ok {
		return -1, errors.New("didn't find responseHeader")
	}

	status, ok := responseHeader["status"].(float64)
	if !ok {
		return -1, errors.New("didn't find status")
	}

	if int(status) != 0 {
		msg, ok := responseBody["message"].(string)
		if !ok {
			return -1, errors.New("no msg found in error message")

		}
		return -1, errors.New(fmt.Sprintf("Error: %v with code %d", msg, int(status)))
	}
	return int(status), nil
}

func (sc *SLClient) DecodeCollectionHealth(responseBody map[string]interface{}) error {
	clusterInfo, ok := responseBody["cluster"].(map[string]interface{})
	if !ok {
		return errors.New("didn't find cluster")
	}
	collections, ok := clusterInfo["collections"].(map[string]interface{})
	if !ok {
		return errors.New("didn't find collections")
	}
	for name, info := range collections {
		collectionInfo := info.(map[string]interface{})
		health, ok := collectionInfo["health"].(string)
		if !ok {
			return errors.New("didn't find health")
		}
		if health != "GREEN" {
			sc.Config.log.Error(errors.New(""), fmt.Sprintf("Health of collection %s IS NOT GREEN", name))
			return errors.New(fmt.Sprintf("health for collection %s is not green", name))
		}
	}
	return nil
}

func (sc *SLClient) GetCollectionList(responseBody map[string]interface{}) ([]string, error) {
	collectionList, ok := responseBody["collections"].([]interface{})
	if !ok {
		return []string{}, errors.New("didn't find collection list")
	}

	collections := make([]string, 0)

	for idx := range collectionList {
		collections = append(collections, collectionList[idx].(string))
	}
	return collections, nil
}

func (sc *SLClient) SearchCollection(collections []string) bool {
	for _, collection := range collections {
		if collection == "kubedb-collection" {
			return true
		}
	}
	return false
}
