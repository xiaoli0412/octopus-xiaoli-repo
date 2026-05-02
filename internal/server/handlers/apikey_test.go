package handlers

import (
	"net/http"
	"testing"
)

func TestDeleteAPIKeyRejectsNonPositivePathIDs(t *testing.T) {
	setupHandlerTest(t)

	for _, id := range []string{"0", "-1"} {
		recorder := performParamHandlerRequest(t, http.MethodDelete, "/api/v1/apikey/delete/"+id, nil, map[string]string{"id": id}, deleteAPIKey)
		if recorder.Code != http.StatusBadRequest {
			t.Fatalf("id %s status = %d, want %d, body = %s", id, recorder.Code, http.StatusBadRequest, recorder.Body.String())
		}
		res := decodeHandlerResponse(t, recorder)
		if res.Message != "Invalid parameter" {
			t.Fatalf("id %s message = %q, want %q", id, res.Message, "Invalid parameter")
		}
	}
}
