package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
)

type BillingData struct {
	TotalUsage float64 `json:"total_usage"`
}

type SubscriptionData struct {
	HardLimitUSD float64 `json:"hard_limit_usd"`
}

type ModelData struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

type CompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func main() {
	a := app.New()
	w := a.NewWindow("API工具")
	w.Resize(fyne.NewSize(600, 600))

	apiURL := widget.NewEntry()
	apiURL.SetPlaceHolder("请输入API URL")

	apiKey := widget.NewEntry()
	apiKey.SetPlaceHolder("请输入您的API Key")

	output := widget.NewMultiLineEntry()
	output.SetPlaceHolder("输出将在这里展示...\n")
	output.SetMinRowsVisible(15) // 确保显示至少15行

	outputContainer := container.NewMax(output) // 使用Max容器让多行文本区撑满

	getBalance := widget.NewButton("获取余额", func() {
		if err := validateInputs(apiURL.Text, apiKey.Text); err != nil {
			dialog.ShowError(err, w)
			return
		}

		url := fmt.Sprintf("%s/v1/dashboard/billing/subscription", strings.TrimSpace(apiURL.Text))
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", strings.TrimSpace(apiKey.Text)))
		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			dialog.ShowError(fmt.Errorf("请求失败: %v", err), w)
			return
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		var data SubscriptionData
		if err := json.Unmarshal(body, &data); err != nil {
			dialog.ShowError(fmt.Errorf("解析响应失败: %v", err), w)
			return
		}

		startDate := "2021-01-01"
		endDate := "2022-01-01"

		billingURL := fmt.Sprintf("%s/v1/dashboard/billing/usage?start_date=%s&end_date=%s", strings.TrimSpace(apiURL.Text), startDate, endDate)
		req, _ = http.NewRequest("GET", billingURL, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", strings.TrimSpace(apiKey.Text)))

		resp, err = client.Do(req)
		if err != nil {
			dialog.ShowError(fmt.Errorf("请求失败: %v", err), w)
			return
		}
		defer resp.Body.Close()

		billingBody, _ := ioutil.ReadAll(resp.Body)
		var billingData BillingData
		if err := json.Unmarshal(billingBody, &billingData); err != nil {
			dialog.ShowError(fmt.Errorf("解析响应失败: %v", err), w)
			return
		}

		remaining := data.HardLimitUSD - billingData.TotalUsage/100
		text := fmt.Sprintf("总额: %.4f USD\n已用: %.4f USD\n剩余: %.4f USD", data.HardLimitUSD, billingData.TotalUsage/100, remaining)
		output.SetText(text)
	})

	getModels := widget.NewButton("获取模型列表", func() {
		if err := validateInputs(apiURL.Text, apiKey.Text); err != nil {
			dialog.ShowError(err, w)
			return
		}

		url := fmt.Sprintf("%s/v1/models", strings.TrimSpace(apiURL.Text))
		req, _ := http.NewRequest("GET", url, nil)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", strings.TrimSpace(apiKey.Text)))

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			dialog.ShowError(fmt.Errorf("请求失败: %v", err), w)
			return
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		var modelData ModelData
		if err := json.Unmarshal(body, &modelData); err != nil {
			dialog.ShowError(fmt.Errorf("解析响应失败: %v", err), w)
			return
		}

		models := make([]string, len(modelData.Data))
		for i, model := range modelData.Data {
			models[i] = model.ID
		}
		text := fmt.Sprintf("模型列表:\n%s", strings.Join(models, "\n"))
		output.SetText(text)
	})

	testModel := widget.NewButton("测试模型", func() {
		modelNameEntry := widget.NewEntry()
		modelNameEntry.SetPlaceHolder("模型名称 (gpt-3.5-turbo)")

		fullResponseCheckbox := widget.NewCheck("返回完整信息", nil)

		var modal *widget.PopUp
		submitButton := widget.NewButton("提交", func() {
			modelName := modelNameEntry.Text
			if modelName == "" {
				modelName = "gpt-3.5-turbo"
			}
			testModelRequest(apiURL.Text, apiKey.Text, modelName, fullResponseCheckbox.Checked, output, w)
			if modal != nil {
				modal.Hide()
			}
		})

		content := container.NewVBox(
			widget.NewLabel("请输入模型名称 (默认使用 gpt-3.5-turbo):"),
			modelNameEntry,
			fullResponseCheckbox,
		)

		closeButton := widget.NewButton("关闭", func() {
			if modal != nil {
				modal.Hide()
			}
		})

		content.Add(container.NewHBox(submitButton, closeButton))

		modal = widget.NewModalPopUp(content, w.Canvas())
		modal.Resize(fyne.NewSize(300, 200))
		modal.Show()
	})

	layout := container.NewVBox(
		widget.NewLabel("API URL:"),
		apiURL,
		widget.NewLabel("API Key:"),
		apiKey,
		getBalance,
		getModels,
		testModel,
		outputContainer, // 使用指定大小的容器
	)

	w.SetContent(layout)
	w.ShowAndRun()
}

func validateInputs(apiURL, apiKey string) error {
	if apiURL == "" || apiKey == "" {
		return fmt.Errorf("请填写API URL和API Key")
	}
	return nil
}

func testModelRequest(apiURL, apiKey, modelName string, fullResponse bool, output *widget.Entry, w fyne.Window) {
	if err := validateInputs(apiURL, apiKey); err != nil {
		dialog.ShowError(err, w)
		return
	}

	url := fmt.Sprintf("%s/v1/chat/completions", strings.TrimSpace(apiURL))
	data := map[string]interface{}{
		"model": modelName,
		"messages": []map[string]string{
			{"role": "user", "content": "say this is a test!"},
		},
	}
	jsonData, _ := json.Marshal(data)

	req, _ := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", strings.TrimSpace(apiKey)))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		dialog.ShowError(fmt.Errorf("请求失败: %v", err), w)
		return
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	var completionResponse CompletionResponse
	if err := json.Unmarshal(body, &completionResponse); err != nil {
		dialog.ShowError(fmt.Errorf("解析响应失败: %v", err), w)
		return
	}

	if fullResponse {
		var prettyJSON bytes.Buffer
		error := json.Indent(&prettyJSON, body, "", "  ")
		if error != nil {
			output.SetText(fmt.Sprintf("格式化JSON失败: %v", error))
		} else {
			output.SetText(prettyJSON.String())
		}
	} else if len(completionResponse.Choices) > 0 {
		output.SetText(completionResponse.Choices[0].Message.Content)
	} else {
		output.SetText("未收到模型回应")
	}
}
