import tkinter as tk
from tkinter import scrolledtext, StringVar, OptionMenu, messagebox
import requests
import json

def send_request():
    if not url_entry.get():
        messagebox.showerror("Error", "请填写您的API URL")
        return
    clear_response_text_area()
    url = url_entry.get()
    if not url.endswith('/v1/chat/completions'):
        url += '/v1/chat/completions'
    
    model = model_var.get()
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
        'Authorization': f'Bearer {auth_entry.get()}'
    }
    
    response = requests.post(url, headers=headers, json=payload)
    response_text = response.text
    pretty_json = json.dumps(json.loads(response_text), indent=4)
    response_text_area.config(state=tk.NORMAL)
    response_text_area.insert(tk.END, pretty_json)
    response_text_area.config(state=tk.DISABLED)
    
    if response.status_code == 200:
        messagebox.showinfo("成功", "测试成功")
    else:
        messagebox.showerror("错误", f"请求失败: {response.status_code} {response.text}")

def get_models():
    if not url_entry.get():
        messagebox.showerror("Error", "请填写您的API URL")
        return
    clear_response_text_area()
    url = url_entry.get()
    if not url.endswith('/v1/models'):
        url += '/v1/models'
    
    headers = {
        'Authorization': f'Bearer {auth_entry.get()}'
    }
    
    response = requests.get(url, headers=headers)
    
    if response.status_code == 200:
        data = response.json()
        model_ids = [model['id'] for model in data['data']]
        model_list = ",\n".join(model_ids)
        response_text_area.config(state=tk.NORMAL)
        response_text_area.insert(tk.END, model_list)
        response_text_area.config(state=tk.DISABLED)
        
        # Update the model dropdown
        model_var.set('')
        model_dropdown['menu'].delete(0, 'end')
        for model_id in model_ids:
            model_dropdown['menu'].add_command(label=model_id, command=tk._setit(model_var, model_id))
        
        messagebox.showinfo("成功", "已成功获取模型")
    else:
        response_text_area.config(state=tk.NORMAL)
        response_text_area.insert(tk.END, f"Error: {response.status_code} {response.text}")
        response_text_area.config(state=tk.DISABLED)
        messagebox.showerror("错误", f"获取模型失败: {response.status_code} {response.text}")

def clear_response_text_area():
    response_text_area.config(state=tk.NORMAL)
    response_text_area.delete(1.0, tk.END)
    response_text_area.config(state=tk.DISABLED)

# 创建主窗口
root = tk.Tk()
root.title("API 测试")

# URL输入框
url_label = tk.Label(root, text="API URL:")
url_label.grid(row=0, column=0, padx=10, pady=10, sticky='e')
url_entry = tk.Entry(root, width=50)
url_entry.grid(row=0, column=1, padx=10, pady=10, sticky='w')

# Authorization输入框
auth_label = tk.Label(root, text="API 令牌:")
auth_label.grid(row=1, column=0, padx=10, pady=10, sticky='e')
auth_entry = tk.Entry(root, width=50)
auth_entry.grid(row=1, column=1, padx=10, pady=10, sticky='w')

# 模型选择下拉列表
model_var = StringVar()
model_label = tk.Label(root, text="选择模型:")
model_label.grid(row=2, column=0, padx=10, pady=10, sticky='e')
model_dropdown = OptionMenu(root, model_var, '')
model_dropdown.grid(row=2, column=1, padx=10, pady=10, sticky='w')

# 发送请求按钮
send_button = tk.Button(root, text="开始测试", command=send_request)
send_button.grid(row=3, column=0, columnspan=2, pady=10, sticky='nsew')
send_button.grid_propagate(False)
send_button.config(width=int(root.winfo_width() * 0.8))

# 获取模型按钮
get_models_button = tk.Button(root, text="获取模型", command=get_models)
get_models_button.grid(row=4, column=0, columnspan=2, pady=10, sticky='nsew')
get_models_button.grid_propagate(False)
get_models_button.config(width=int(root.winfo_width() * 0.8))

# 响应文本框
response_text_area = scrolledtext.ScrolledText(root, width=60, height=15, state=tk.DISABLED)
response_text_area.grid(row=5, column=0, columnspan=2, padx=10, pady=10, sticky='nsew')
response_text_area.grid_propagate(False)
response_text_area.config(width=int(root.winfo_width() * 0.8))

# 设置列和行的权重以实现自适应
root.columnconfigure(0, weight=1)
root.columnconfigure(1, weight=3)
for i in range(6):
    root.rowconfigure(i, weight=1)

# 运行主循环
root.mainloop()
