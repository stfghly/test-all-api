import PySimpleGUI as sg
import requests
import json

def send_request():
    if not values['-URL-']:
        sg.popup_error("Error", "请填写您的API URL")
        return
    clear_response_text_area()
    url = values['-URL-']
    if not url.endswith('/v1/chat/completions'):
        url += '/v1/chat/completions'
    
    model = values['-MODEL-']
    if not model:
        model = "gpt-3.5-turbo"
    
    payload = {
        "model": model,
        "messages": [
            {
                "role": "user",
                "content": "Hello!"
            }
        ]
    }
    
    headers = {
        'Accept': 'application/json',
        'User-Agent': 'Apifox/1.0.0 (https://apifox.com)',
        'Content-Type': 'application/json',
        'Authorization': f'Bearer {values["-AUTH-"]}'
    }
    
    response = requests.post(url, headers=headers, json=payload)
    response_text = response.text
    pretty_json = json.dumps(json.loads(response_text), indent=4)
    window['-RESPONSE-'].update(pretty_json)
    sg.popup("测试成功")

def get_models():
    if not values['-URL-']:
        sg.popup_error("Error", "请填写您的API URL")
        return
    clear_response_text_area()
    url = values['-URL-']
    if not url.endswith('/v1/models'):
        url += '/v1/models'
    
    headers = {
        'Authorization': f'Bearer {values["-AUTH-"]}'
    }
    
    response = requests.get(url, headers=headers)
    
    if response.status_code == 200:
        data = response.json()
        model_ids = [model['id'] for model in data['data']]
        model_list = ",\n".join(model_ids)
        window['-RESPONSE-'].update(model_list)
        
        # Update the model dropdown
        window['-MODEL-'].update(values=model_ids)
        sg.popup("已成功获取模型")
    else:
        window['-RESPONSE-'].update(f"Error: {response.status_code} {response.text}")

def clear_response_text_area():
    window['-RESPONSE-'].update('')

# 创建主窗口布局
layout = [
    [sg.Text("API URL:"), sg.Input(key='-URL-', size=(50, 1))],
    [sg.Text("API 令牌:"), sg.Input(key='-AUTH-', size=(50, 1))],
    [sg.Text("选择模型:"), sg.Combo([], key='-MODEL-', size=(48, 1))],
    [sg.Button("开始测试"), sg.Button("获取模型")],
    [sg.Text("响应:"), sg.Multiline(key='-RESPONSE-', size=(60, 15), disabled=True)]
]

# 创建窗口
window = sg.Window("API 测试", layout)

# 事件循环
while True:
    event, values = window.read()
    if event == sg.WIN_CLOSED:
        break
    elif event == "开始测试":
        send_request()
    elif event == "获取模型":
        get_models()

window.close()
