package main

import (
	"bytes"
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/getsentry/sentry-go"
	"github.com/logrusorgru/aurora/v4"
	"log"
	"net/http"
)

const QueryDiscordID = `SELECT "user".discord_id FROM "user" WHERE "user".wii_id = $1 LIMIT 1`

func NewResponse(r *http.Request, w *http.ResponseWriter, xmlType XMLType) *Response {
	return &Response{
		ResponseFields: KVFieldWChildren{
			XMLName: xml.Name{Local: "response"},
			Value:   nil,
		},
		wiiID:               r.Header.Get("X-WiiID"),
		request:             r,
		writer:              w,
		isMultipleRootNodes: xmlType == 1,
	}
}

// AddCustomType adds a given key by name to a specified structure.
func (r *Response) AddCustomType(customType any) {
	k, ok := r.ResponseFields.(KVFieldWChildren)
	if ok {
		k.Value = append(k.Value, customType)
		r.ResponseFields = k
		return
	}

	// Now check if the fields is an array of any.
	array, ok := r.ResponseFields.([]any)
	if ok {
		r.ResponseFields = append(r.ResponseFields.([]any), array)
	}
}

// AddKVNode adds a given key by name to a specified value, such as <key>value</key>.
func (r *Response) AddKVNode(key string, value string) {
	k, ok := r.ResponseFields.(KVFieldWChildren)
	if !ok {
		return
	}

	k.Value = append(k.Value, KVField{
		XMLName: xml.Name{Local: key},
		Value:   value,
	})

	r.ResponseFields = k
}

// AddKVWChildNode adds a given key by name to a specified value, such as <key><child>...</child></key>.
func (r *Response) AddKVWChildNode(key string, value any) {
	k, ok := r.ResponseFields.(KVFieldWChildren)
	if !ok {
		return
	}

	k.Value = append(k.Value, KVFieldWChildren{
		XMLName: xml.Name{Local: key},
		Value:   []any{value},
	})

	r.ResponseFields = k
}

func (r *Response) toXML() (string, error) {
	var contents string

	if r.isMultipleRootNodes {
		var temp []byte
		var err error
		array, ok := r.ResponseFields.([]any)
		if ok {
			for _, a := range array {
				temp, err = xml.MarshalIndent(a, "", "  ")
				if err != nil {
					return "", err
				}

				contents += string(temp) + "\n"
			}
		} else {
			temp, err = xml.MarshalIndent(r.ResponseFields, "", "  ")
			if err != nil {
				return "", err
			}

			contents += string(temp) + "\n"
		}

		// Now the version and API tags
		version, apiStatus := GenerateVersionAndAPIStatus()
		temp, err = xml.MarshalIndent(version, "", "  ")
		if err != nil {
			return "", err
		}

		contents += string(temp) + "\n"

		temp, err = xml.MarshalIndent(apiStatus, "", "  ")
		if err != nil {
			return "", err
		}

		contents += string(temp)
	} else {
		version, apiStatus := GenerateVersionAndAPIStatus()
		r.AddCustomType(version)
		r.AddCustomType(apiStatus)
		temp, err := xml.MarshalIndent(r.ResponseFields, "", "  ")
		if err != nil {
			return "", err
		}

		contents += string(temp)
	}

	return contents, nil
}

func GenerateVersionAndAPIStatus() (*KVField, *KVFieldWChildren) {
	version := KVField{
		XMLName: xml.Name{Local: "version"},
		Value:   "1",
	}

	apiStatus := KVFieldWChildren{
		XMLName: xml.Name{Local: "apiStatus"},
		Value: []any{
			KVField{
				XMLName: xml.Name{Local: "code"},
				Value:   "0",
			},
		},
	}

	return &version, &apiStatus
}

// BoolToInt converts a boolean value to an integer.
// This is needed because Nintendo wants the integer, not the string literal.
func BoolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func PostDiscordWebhook(title, message, url string, color int) {
	theMap := map[string]any{
		"content": nil,
		"embeds": []map[string]any{
			{
				"title":       title,
				"description": message,
				"color":       color,
			},
		},
	}

	jsonData, _ := json.Marshal(theMap)
	_, _ = http.Post(url, "application/json", bytes.NewBuffer(jsonData))
}

// ReportError helps make errors nicer. First it logs the error to Sentry,
// then writes a response for the server to send.
func (r *Response) ReportError(err error, code int, response *map[string]any) {
	sentry.CaptureException(err)
	log.Printf("An error has occurred: %s", aurora.Red(err.Error()))

	var discord_id string
	row := pool.QueryRow(context.Background(), QueryDiscordID, r.request.Header.Get("X-WiiID"))
	_err := row.Scan(&discord_id)

	// Re-marshal back into string
	var b []byte
	if response != nil {
		b, _ = json.Marshal(*response)
	}

	sentry.WithScope(func(s *sentry.Scope) {
		s.SetTag("Discord ID", discord_id)
		if b != nil {
			s.SetExtra("Response", string(b))
		}

		sentry.CaptureException(_err)
	})

	errorString := fmt.Sprintf("%s\nWii ID: %s\nDiscord ID: %s", err.Error(), r.wiiID, discord_id)
	PostDiscordWebhook("An error has occurred in Demae Domino's!", errorString, config.ErrorWebhook, 16711711)

	// Write response
	r.hasError = true
	http.Error(*r.writer, err.Error(), code)
}

func printError(w http.ResponseWriter, reason string, code int) {
	http.Error(w, reason, code)
	log.Print("Failed to handle request: ", aurora.Red(reason))
}
