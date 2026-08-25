package es

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const (
	UrlBase       = "https://es.gaido.dev"
	authorization = "Basic ZWxhc3RpYzpOWjkqdVhDTGJDUmV1MTZ0Y0VHcw=="
)

type D map[string]any
type A []any

func putHeaders(r *http.Request) {
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Authorization", authorization)
}

func do(method string, ed string, values url.Values, in any, out any) error {
	u := fmt.Sprintf("%v%v", UrlBase, ed)
	if values != nil {
		u += "?" + values.Encode()
	}
	var (
		b   []byte
		err error
		req *http.Request
	)
	if in != nil {
		b, err = json.Marshal(in)
		if err != nil {
			return err
		}
		req, err = http.NewRequest(method, u, bytes.NewReader(b))
	} else {
		req, err = http.NewRequest(method, u, nil)
	}
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authorization)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode >= 400 {
		b, _ = io.ReadAll(res.Body)
		return fmt.Errorf("status code: %d, body: %s", res.StatusCode, string(b))
	}
	b, err = io.ReadAll(res.Body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, out); err != nil {
		return fmt.Errorf(err.Error(), string(b))
	}
	return nil
}

type Hit struct {
	Index string   `json:"_index"`
	Type  string   `json:"_type"`
	Id    string   `json:"_id"`
	Score *float64 `json:"_score"`
	Sort  []any    `json:"sort"`
}

type Index struct {
	Index string `json:"_index"`
	Id    string `json:"_id,omitempty"`
}

type Meta struct {
	Index Index `json:"index"`
}

type Update struct {
	Index string `json:"_index"`
	Id    string `json:"_id"`
}

type UpdateMeta struct {
	Update Update `json:"update"`
}

func BulkIndex(index string, docs []any) error {
	u := fmt.Sprintf("%v/_bulk", UrlBase)
	var buf bytes.Buffer
	var b []byte
	for _, doc := range docs {
		meta := Meta{Index: Index{
			Index: index,
		}}
		b, _ = json.Marshal(meta)
		buf.Write(b)
		buf.WriteString("\n")
		if s, ok := doc.(string); ok {
			buf.WriteString(s)
		} else {
			b, _ = json.Marshal(doc)
			buf.Write(b)
		}
		buf.WriteString("\n")
	}
	req, err := http.NewRequest("POST", u, &buf)
	if err != nil {
		return fmt.Errorf("es bulk index: %w", err)
	}
	putHeaders(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("es bulk index: %w, %v", err, string(b))
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return fmt.Errorf("es bulk index: status %v, %v", res.StatusCode, string(b))
	}
	res.Body.Close()
	return nil
}

func BulkIndexes(indexes []string, docs []any) error {
	if len(indexes) != len(docs) {
		return fmt.Errorf("es bulk indexes: len(indexes) != len(docs)")
	}
	u := fmt.Sprintf("%v/_bulk", UrlBase)
	var buf bytes.Buffer
	var b []byte
	for i, doc := range docs {
		meta := Meta{Index: Index{
			Index: indexes[i],
		}}
		b, _ = json.Marshal(meta)
		b2, _ := json.Marshal(doc)
		buf.Write(b)
		buf.WriteString("\n")
		buf.Write(b2)
		buf.WriteString("\n")
	}
	req, err := http.NewRequest("POST", u, &buf)
	if err != nil {
		return fmt.Errorf("es bulk indexes: %w", err)
	}
	putHeaders(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("es bulk indexes: %w, %v", err, string(b))
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return fmt.Errorf("es bulk indexes: status %v, %v", res.StatusCode, string(b))
	}
	res.Body.Close()
	return nil
}

func BulkUpdate(index string, ids []string, docs []any) error {
	u := fmt.Sprintf("%v/_bulk", UrlBase)
	var buf bytes.Buffer
	var b []byte
	for i := range ids {
		id := ids[i]
		doc := docs[i]
		update := UpdateMeta{Update: Update{
			Index: index,
			Id:    id,
		}}
		b, _ = json.Marshal(update)
		buf.Write(b)
		buf.WriteString("\n")
		b, _ = json.Marshal(doc)
		buf.Write(b)
		buf.WriteString("\n")
	}
	req, err := http.NewRequest("POST", u, &buf)
	if err != nil {
		return fmt.Errorf("es bulk update: %w", err)
	}
	putHeaders(req)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("es bulk update: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return fmt.Errorf("es bulk update: status %v", res.StatusCode)
	}
	res.Body.Close()
	return nil

}
