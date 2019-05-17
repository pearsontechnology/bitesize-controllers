package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"io/ioutil"
	"strings"
)

func main() {
	sn := "ingress/cnn.com"
	svc := session.Must(session.NewSession())
	sm := secretsmanager.New(svc, aws.NewConfig().WithRegion("us-west-2"))
	_, err := sm.GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: &sn})
	if err != nil {
		panic(err.Error())
	}
	//   fmt.Println(*output.SecretString)

	input := &secretsmanager.ListSecretsInput{}

	result, err := sm.ListSecrets(input)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			switch aerr.Code() {
			case secretsmanager.ErrCodeInvalidParameterException:
				fmt.Println(secretsmanager.ErrCodeInvalidParameterException, aerr.Error())
			case secretsmanager.ErrCodeInvalidNextTokenException:
				fmt.Println(secretsmanager.ErrCodeInvalidNextTokenException, aerr.Error())
			case secretsmanager.ErrCodeInternalServiceError:
				fmt.Println(secretsmanager.ErrCodeInternalServiceError, aerr.Error())
			default:
				fmt.Println(aerr.Error())
			}
		} else {
			// Print the error, cast err to awserr.Error to get the Code and
			// Message from an error.
			fmt.Println(err.Error())
		}
		return
	}

	//    fmt.Println(result)
	for _, s := range result.SecretList {
		if strings.HasPrefix(*s.Name, "ingress/c") {
			fmt.Println(*s.Name)

			output, err := sm.GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: s.Name})
			if err != nil {
				panic(err.Error())
			}
			//			fmt.Printf("%v+\n",output)
			//			fmt.Println(*output.SecretString)
			fmt.Printf("%T\n", *output.SecretString)
			var raw map[string]string
			json.Unmarshal([]byte(*output.SecretString), &raw)
			out, _ := json.Marshal(raw)
			println(string(out))
			fmt.Printf("%v+\n", raw)
			fmt.Println(raw["key"])
			rawString, err := base64.StdEncoding.DecodeString(raw["key"])
			if err != nil {
				panic(err)
			}

			fmt.Printf("%s\n", rawString)
			filename := "/tmp/" + *s.Name + ".key"
			fmt.Println(filename)

			err = ioutil.WriteFile(filename, rawString, 0600)
			if err != nil {
				panic(err)
			}
		}
	}
}
