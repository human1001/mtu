package com

import (
	"bytes"
	"io/ioutil"
	"unicode/utf8"

	"github.com/gogs/chardet"
	"golang.org/x/net/html/charset"
)

// ToUtf8 转为任何编码(尽可能)为utf8编码
func ToUtf8(s []byte) []byte {

	// chardet echo charsets:Shift_JIS,EUC-JP,EUC-KR,Big5,GB18030,ISO-8859-2(windows-1250),ISO-8859-5,ISO-8859-6,ISO-8859-7,indows-1253,ISO-8859-8(windows-1255),ISO-8859-8-I,ISO-8859-9(windows-1254),windows-1256,windows-1251,KOI8-R,IBM424_rtl,IBM424_ltr,IBM420_rtl,IBM420_ltr,ISO-2022-JP

	d := chardet.NewTextDetector() //charset.DetermineEncoding 不是很准
	var rs *chardet.Result
	var err error
	if len(s) > 1024 {
		if utf8.Valid(s[:1024]) {
			return s
		}
		rs, err = d.DetectBest(s[:1024])
	} else {
		if utf8.Valid(s) {
			return s
		}
		rs, err = d.DetectBest(s)
	}
	if Errorlog(err) {
		// for gbk
		return nil
	}

	// 转换
	// charset input charsets:utf-8,ibm866,iso-8859-2,iso-8859-3,iso-8859-4,iso-8859-5,iso-8859-6,iso-8859-7,iso-8859-8,iso-8859-8-i,iso-8859-10,iso-8859-13,iso-8859-14,iso-8859-15,iso-8859-16,koi8-r,koi8-u,macintosh,windows-874,windows-1250,windows-1251,windows-1252,windows-1253,windows-1254,windows-1255,windows-1256,windows-1257,windows-1258,x-mac-cyrillic,gbk,gb18030,big5,euc-jp,iso-2022-jp,shift_jis,euc-kr,replacement,utf-16be,utf-16le,x-user-defined,

	var maps map[string]string = make(map[string]string)
	maps = map[string]string{
		"Shift_JIS":    "shift_jis",
		"EUC-JP":       "euc-jp",
		"EUC-KR":       "euc-kr",
		"Big5":         "big5",
		"GB18030":      "gb18030",
		"ISO-8859-2 ":  "iso-8859-2",
		"ISO-8859-5":   "iso-8859-5",
		"ISO-8859-6":   "iso-8859-6",
		"ISO-8859-7":   "iso-8859-7",
		"ISO-8859-8":   "iso-8859-8",
		"ISO-8859-8-I": "iso-8859-8-i",
		"ISO-8859-9":   "iso-8859-10",
		"windows-1256": "windows-1256",
		"windows-1251": "windows-1251",
		"KOI8-R":       "koi8-r",
		"ISO-2022-JP":  "iso-2022-jp",
		"UTF-16BE ":    "utf-16be",
		"UTF-16LE ":    "utf-16le",
	}

	ct := maps[rs.Charset]
	if ct == "" { // 使用 charset.DetermineEncoding
		_, name, b := charset.DetermineEncoding([]byte(s), "utf-8")
		if b {
			return s
		}
		ct = name
	}

	byteReader := bytes.NewReader(s)
	reader, err1 := charset.NewReaderLabel(ct, byteReader)
	r, err2 := ioutil.ReadAll(reader)
	if Errorlog(err1, err2) {
		return nil
	}
	return r
}
