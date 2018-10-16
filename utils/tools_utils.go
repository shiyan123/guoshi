package utils

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/tealeg/xlsx"
	"io"
	"math/rand"
	"reflect"
	"sort"
)

//map key 字典排序 并 md5hex
func DictionaryOrderSign(paramMap map[string]string, baseUrl string) (sign string, originStr string) {
	secretKey := paramMap["secretKey"]
	ks := make([]string, 0, len(paramMap))
	delete(paramMap, "secretKey")

	for k := range paramMap {
		if k == "sign" {
			continue
		}
		ks = append(ks, k)
	}
	sort.Strings(ks)

	Str := ""
	for m, k := range ks {
		v := paramMap[k]
		if m == 0 {
			Str = fmt.Sprintf("%s%s=%s", Str, k, v)
		} else {
			Str = fmt.Sprintf("%s&%s=%s", Str, k, v)
		}

	}
	fmt.Printf("Str: %s\n", Str)
	//拼接地址
	originStr = Str
	Str = fmt.Sprintf("GET%s?%s", baseUrl, Str)
	fmt.Printf("last str : %s \n", Str)

	h := sha1.New()
	io.WriteString(h, Str)
	fmt.Printf("%x\n", h.Sum(nil))
	key := []byte(secretKey)
	mac := hmac.New(sha1.New, key)
	mac.Write([]byte(Str))
	encodeString := base64.StdEncoding.EncodeToString(mac.Sum(nil))
	fmt.Printf("encoding str: %s \n", encodeString)
	return encodeString, originStr
}

const chars = "0123456789"

func GetNonce() (nonce string) {
	for i := 0; i < 5; i++ {
		nonce += string(chars[rand.Int()%len(chars)])
	}
	return nonce
}

func GeneralXlsxGenerator(doc interface{}) (body *bytes.Buffer, err error) {
	switch reflect.TypeOf(doc).Kind() {
	case reflect.Slice:
		v := reflect.ValueOf(doc)

		var (
			xFile *xlsx.File
			sheet *xlsx.Sheet
			row   *xlsx.Row
			cell  *xlsx.Cell
		)
		xFile = xlsx.NewFile()
		sheet, err = xFile.AddSheet("sheet1")
		if err != nil {
			return
		}

		for i := 0; i < v.Len(); i++ {
			tt := reflect.TypeOf(v.Index(i).Elem().Interface())
			if i == 0 {
				row = sheet.AddRow()
				for j := 0; j < tt.NumField(); j++ {
					val, ok := tt.Field(j).Tag.Lookup("xson")
					if !ok {
						val = tt.Field(j).Name
					}
					if val == "-" {
						continue
					}
					cell = row.AddCell()
					cell.SetValue(val)
				}
			}

			row = sheet.AddRow()
			vv := reflect.ValueOf(v.Index(i).Elem().Interface())
			for k := 0; k < tt.NumField(); k++ {
				val, _ := tt.Field(k).Tag.Lookup("xson")
				if val == "-" {
					continue
				}
				var a interface{}
				a = vv.Field(k).Interface()
				cell = row.AddCell()
				cell.SetValue(a)
			}
		}

		body = new(bytes.Buffer)
		err = xFile.Write(body)

		//case reflect.Struct:
		//	GeneralXlsxGenerator([]interface{}{doc})
		//case reflect.Map:
		//	newDoc := doc.(map[string]interface{})
		//	slice := make([]interface{}, 0)
		//	for _, v := range newDoc {
		//		slice = append(slice, v)
		//	}
		//	GeneralXlsxGenerator(slice)
	default:
		err = errors.New("invalid type")
	}

	return
}
