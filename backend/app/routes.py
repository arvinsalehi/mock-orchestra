from fastapi import APIRouter, HTTPException
from .models import SessionStartRequest
from .database import db, redis_client
from .connection import manager
from .mqtt_handler import mqtt
import uuid
import json
from datetime import datetime

router = APIRouter()

@router.post("/start-test")
async def start_test(req: SessionStartRequest):
    # 1. Check Redis for EXISTING session
    redis_key = f"session:{req.build_number}"
    active_session = await redis_client.hgetall(redis_key)
    
    if active_session:
        status = active_session.get("status")
        
        # elapsed time for display
        total_seconds = float(active_session.get("total_active_seconds", 0))
        
        # CASE A: Session is ACTIVE -> Return existing session (Login)
        if status == "active":
            # Add time elapsed since last start
            last_start = float(active_session.get("last_active_start", 0))
            if last_start > 0:
                current_duration = datetime.utcnow().timestamp() - last_start
                total_seconds += current_duration

            return {
                "session_hash": active_session.get("session_hash"),
                "build_number": req.build_number,
                "status": "active",
                "operator": active_session.get("operator"),
                "active_time_elapsed": total_seconds,
                "message": "Reconnected to active session."
            }
            
        # CASE B: Session is PENDING (Paused) -> Return info
        if status == "pending":
            return {
                "session_hash": active_session.get("session_hash"),
                "build_number": req.build_number,
                "status": "pending",
                "operator": active_session.get("operator"),
                "active_time_elapsed": total_seconds,
                "message": "Session is currently pending/paused."
            }

    # 2. Redis Miss - New Session Logic
    build_doc = await db.build_numbers.find_one({"build_number": req.build_number})
    if not build_doc:
        raise HTTPException(status_code=404, detail="Build number not found")

    # 4. Create NEW Active Session
    session_hash = str(uuid.uuid4())
    start_time = datetime.utcnow().isoformat()
    
    session_data = {
        "session_hash": session_hash,
        "build_number": req.build_number,
        "operator": req.operator,
        "status": "active",
        "start_time": start_time,
        "last_active_start": datetime.utcnow().timestamp(), 
        "total_active_seconds": 0
    }

    await redis_client.hset(redis_key, mapping=session_data)
    await redis_client.hset(f"session_hash:{session_hash}", mapping=session_data) 
    await redis_client.expire(redis_key, 86400)

    mqtt.publish(f"hub/{session_hash}/start", json.dumps(session_data))

    return session_data

@router.patch("/pause/{session_hash}")
async def pause_session(session_hash: str):
    # Get metadata
    meta_key = f"session_hash:{session_hash}"
    meta = await redis_client.hgetall(meta_key)
    if not meta: raise HTTPException(status_code=404)
    
    if meta['status'] == 'pending':
        return {"status": "pending", "message": "Already paused"}

    # Calculate active time since last start
    now_ts = datetime.utcnow().timestamp()
    last_start = float(meta.get("last_active_start", now_ts))
    prev_total = float(meta.get("total_active_seconds", 0))
    
    new_total = prev_total + (now_ts - last_start)

    # Update Redis: Set status to pending, update total time, clear active start
    updates = {
        "status": "pending",
        "total_active_seconds": new_total,
        "last_active_start": 0 # Cleared while paused
    }
    
    redis_key = f"session:{meta['build_number']}"
    await redis_client.hset(redis_key, mapping=updates)
    await redis_client.hset(meta_key, mapping=updates) # Sync reverse lookup
    
    # Notify Hub & Close WS
    mqtt.publish(f"tests/{session_hash}/pause", "true")
    await manager.disconnect_all(session_hash)
    
    return {"status": "pending", "active_time_elapsed": new_total}

# Redundant. Will be removed
@router.patch("/resume/{session_hash}")
async def resume_session(session_hash: str):
    meta_key = f"session_hash:{session_hash}"
    meta = await redis_client.hgetall(meta_key)
    if not meta: raise HTTPException(status_code=404)

    if meta['status'] == 'active':
        return {"status": "active", "message": "Already active"}

    # Set new active start time
    updates = {
        "status": "active",
        "last_active_start": datetime.utcnow().timestamp()
    }

    redis_key = f"session:{meta['build_number']}"
    await redis_client.hset(redis_key, mapping=updates)
    await redis_client.hset(meta_key, mapping=updates)

    mqtt.publish(f"tests/{session_hash}/resume", "true")
    
    return {"status": "active"}


@router.patch("/finish/{session_hash}")
async def finish_session(session_hash: str):
    meta = await redis_client.hgetall(f"session_hash:{session_hash}")
    if not meta: raise HTTPException(status_code=404)
    
    # Calculate final duration
    total_seconds = float(meta.get("total_active_seconds", 0))
    if meta['status'] == 'active':
        last_start = float(meta.get("last_active_start", 0))
        if last_start > 0:
            total_seconds += (datetime.utcnow().timestamp() - last_start)
    
    raw_results = await redis_client.lrange(f"tests:{session_hash}", 0, -1)
    results = [json.loads(r) for r in raw_results]

    # 2. Persist to MongoDB
    final_doc = {
        **meta,
        "status": "completed",
        "total_duration_seconds": total_seconds, # Storing the calculated duration
        "completed_at": datetime.now(),
    }
    await db.tests.insert_one(final_doc)

    # 3. Cleanup Redis
    await redis_client.delete(f"session:{meta['build_number']}")
    await redis_client.delete(f"session_hash:{session_hash}")
    await redis_client.delete(f"tests:{session_hash}")

    # 4. Notify & Close
    mqtt.publish(f"tests/{session_hash}/finished", "true")
    await manager.disconnect_all(session_hash)

    return {"status": "completed"}

