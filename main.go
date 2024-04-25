package main

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
	"github.com/tmc/langchaingo/prompts"
)

func main() {
	r := gin.Default()

	v1 := r.Group("/api/v1")

	v1.POST("/translator", translator)

	r.Run(":8888")
}

func translator(c *gin.Context) {
	var requestData struct {
		OutputLang string `json:"outputlang"`
		Text       string `json:"text"`
	}

	if err := c.BindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error:": "绑定json参数错误"})
		return
	}

	// 创建prompt
	prompt := prompts.NewChatPromptTemplate([]prompts.MessageFormatter{
		// 模型处理规则描述
		prompts.NewSystemMessagePromptTemplate("你是一个销售人员态度分析引擎,通过对销售人员的态度进行评分,如果销售人员在销售过程中对顾客有嘲讽等恶意情绪请结合语气给出评分,满分100分,你只需要输出评分和简单评价", nil),
		// 输入
		prompts.NewHumanMessagePromptTemplate(`{{.text}}:{{.outputlang}}`, []string{"text", "outputlang"}),
	})

	// 填充prompt
	value := map[string]any{
		"outputlang": requestData.OutputLang,
		"text":       requestData.Text,
	}

	messages, err := prompt.FormatMessages(value)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error:": "消息处理错误!"})
		return
	}

	//链接ollama
	llm, err := ollama.New(ollama.WithModel("qwen"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error:": err})
	}

	content := []llms.MessageContent{
		llms.TextParts(messages[0].GetType(), messages[0].GetContent()),
		llms.TextParts(messages[1].GetType(), messages[1].GetContent()),
	}

	response, err := llm.GenerateContent(context.Background(), content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error:": err})
	}

	c.JSON(http.StatusOK, gin.H{"response": response.Choices[0].Content})
}
