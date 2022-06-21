package output

import (
	"net/http"

	"github.com/no-mole/neptune/json"
)

var jsonContentType = []string{"application/json; charset=utf-8"}

type nJson struct {
	Data interface{}
}

func (r nJson) Render(w http.ResponseWriter) (err error) {
	if err = WriteJSON(w, r.Data); err != nil {
		panic(err)
	}
	return
}

// WriteContentType (JSON) writes JSON ContentType.
func (r nJson) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}

func WriteJSON(w http.ResponseWriter, obj interface{}) error {
	writeContentType(w, jsonContentType)
	jsonBytes, err := json.MarshalNoTag(obj)
	if err != nil {
		return err
	}
	_, err = w.Write(jsonBytes)
	return err
}

func writeContentType(w http.ResponseWriter, value []string) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = value
	}
}
