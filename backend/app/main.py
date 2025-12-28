from fastapi import FastAPI, WebSocket, WebSocketDisconnect
from fastapi.middleware.cors import CORSMiddleware
from .routes import router
from .mqtt_handler import mqtt
from .connection import manager

app = FastAPI()


app.add_middleware(
    CORSMiddleware,
    allow_origins=["http://localhost:3000", "http://127.0.0.1:3000", "*"],
    allow_methods=["*"],
    allow_headers=["*"],
)

mqtt.init_app(app)

app.include_router(router, prefix="/api")

@app.websocket("/ws/{session_hash}")
async def websocket_endpoint(websocket: WebSocket, session_hash: str):
    await manager.connect(session_hash, websocket)
    try:
        while True:
            await websocket.receive_text() # Keep alive
    except WebSocketDisconnect:
        manager.disconnect(session_hash, websocket)

