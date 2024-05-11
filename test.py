import websocket


def on_message(ws, message):
    print(f'got message {message}\n')


for i in range(1, 9999):
    ws = websocket.WebSocketApp("ws://127.0.0.1:322/ws", on_message=on_message)
    ws.run_forever()
    ws.close()
