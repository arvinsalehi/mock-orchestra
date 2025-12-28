import requests
import sys

# Configuration
API_URL = "http://localhost:8000/api"

def create_session(build_num="1.0.0-alpha"):
    print(f"🔨 Creating new Session for Build: {build_num}...")
    
    try:
        # Adjust payload to match your exact Pydantic model
        payload = {
            "build_number": build_num,
            "operator": "Auto-Script",
        }
        
        response = requests.post(f"{API_URL}/start-test", json=payload)
        response.raise_for_status()
        
        data = response.json()
        session_id = data.get("session_hash") or data.get("id")
        
        print(f"\n✅ Session Created Successfully!")
        print(f"🆔 Session ID: {session_id}")
        print(f"🔗 UI URL:     http://localhost:3000/sessions/{session_id}")
        print(f"📋 Command:    ./scripts/test_ui_update.sh {session_id}")
        
        return session_id
        
    except Exception as e:
        print(f"❌ Failed to create session: {e}")
        return None

if __name__ == "__main__":
    build = sys.argv[1] if len(sys.argv) > 1 else "1.0.0-dev"
    create_session(build)
