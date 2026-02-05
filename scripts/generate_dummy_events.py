import json
import random
import time
import os
from datetime import datetime
import uuid

# Configuration
DATA_DIR = "./data/raw/events"
NUM_EVENTS = 1000
CAMPAIGNS = ["spring_sale", "black_friday", "new_user_promo"]
EVENT_TYPES = ["impression", "click", "add_to_cart", "purchase"]

def generate_event():
    return {
        "event_id": str(uuid.uuid4()),
        "timestamp": datetime.now().isoformat(),
        "campaign_id": random.choice(CAMPAIGNS),
        "event_type": random.choice(EVENT_TYPES),
        "user_id": f"user_{random.randint(1, 100)}",
        "platform": random.choice(["web", "ios", "android"])
    }

def main():
    # Ensure directory exists
    os.makedirs(DATA_DIR, exist_ok=True)
    
    # Create a batch file
    timestamp = int(time.time())
    file_path = f"{DATA_DIR}/events_{timestamp}.json"
    
    print(f"Generating {NUM_EVENTS} events to {file_path}...")
    
    with open(file_path, "w") as f:
        for _ in range(NUM_EVENTS):
            event = generate_event()
            f.write(json.dumps(event) + "\n")
            
    print("Done!")

if __name__ == "__main__":
    main()
