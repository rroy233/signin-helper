package main

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"os"
	"signin/Logger"
	"strings"
)


func viewIndex(c *gin.Context)  {
	authorizationMiddleware(c, 0)
	_, err := getAuthFromContext(c)
	if err != nil {
		middleWareRedirect(c)
		return
	}

	c.Data(200, ContentTypeHTML, views("dist.index"))
}

func viewReg(c *gin.Context) {
	_, err := getAuthFromContext(c)
	if err != nil {
		redirectToLogin(c)
		return
	}
	c.Data(200, ContentTypeHTML, views("reg"))
}


//模板加载函数
func views(template string, params ...map[string]string) (html []byte) {
	name := ""
	data := make([]string, 0)
	if strings.Index(template, ".") != -1 {
		data = strings.Split(template, ".")
		for _, n := range data {
			name = name + "/" + n
		}
	} else {
		name = "/" + template
	}
	file, err := os.Open("./views" + name + ".html")
	defer file.Close()
	if err != nil {
		Logger.FATAL.Println("模板读取失败:", err.Error())
		html = []byte("模板读取失败")
		return
	}
	html, _ = ioutil.ReadAll(file)
	html = bytes.Replace(html, []byte("{{api_url}}"), []byte(config.General.BaseUrl), -1)

	if len(params) != 0 {
		for k, v := range params[0] {
			html = bytes.Replace(html, []byte("{{"+k+"}}"), []byte(v), -1)
		}
	}

	//替换版本
	html = bytes.Replace(html, []byte("{{version}}"), []byte(BackEndVer), -1)

	return
}

func versionHandler(c *gin.Context) {
	res := new(resVersion)
	res.Status = 0
	res.Data.Version = BackEndVer
	c.JSON(200,res)
}
