package sns

import (
	"crypto/x509"
	"encoding/base64"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

var eventTypes = map[string][]string{
	"Notification":             {"Message", "MessageId", "Subject", "Timestamp", "TopicArn", "Type"},
	"SubscriptionConfirmation": {"Message", "MessageId", "SubscribeUrl", "Timestamp", "Token", "TopicArn", "Type"},
	"UnsubscribeConfirmation":  {"Message", "MessageId", "SubscribeUrl", "Timestamp", "Token", "TopicArn", "Type"},
}

func verifySignature(event SnsEvent) error {
	if event.SignatureVersion != "1" {
		return errors.New("invalid signature version != 1") // Only Version 1 is implemented
	}

	x509CertResponse, err := http.Get(event.SigningCertURL)
	if err != nil || x509CertResponse.StatusCode != 200 {
		if err == nil {
			err = errors.New("status code != 200")
		}
		return err
	}

	x509Cert, err := ioutil.ReadAll(x509CertResponse.Body)
	if err != nil {
		return err
	}

	cert, err := x509.ParseCertificate(x509Cert)
	if err != nil {
		return err
	}

	msgSignature, err := createSignature(event)
	if err != nil {
		return err
	}

	snsSignature, err := base64.StdEncoding.DecodeString(event.Signature)
	if err != nil {
		return err
	}

	return cert.CheckSignature(x509.SHA1WithRSA, msgSignature, snsSignature)
}

func createSignature(event SnsEvent) ([]byte, error) {
	pairs, ok := eventTypes[event.Type]
	if !ok {
		return []byte{}, errors.New("invalid event type")
	}

	var res []string
	for _, v := range pairs {
		i := reflect.ValueOf(event) // One day we will rewrite this to use a map instead
		val := reflect.Indirect(i).FieldByName(v)
		if v == "Subject" {
			valStr := val.String()
			if valStr == "" {
				continue
			}
		}
		res = append(res, v, val.String())
	}

	return []byte(strings.Join(res, "\n")), nil
}
