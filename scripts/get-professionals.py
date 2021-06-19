import fileinput
import json
import sys

import requests

def main():
    session = requests.Session()
    for l in fileinput.input():
        l = l.strip()
        try:
            data = fetch(session, l)
            json.dump(data, sys.stdout)
            sys.stdout.write("\n")
            sys.stdout.flush()
        except Exception as e:
            sys.stderr.write(f"{l} error={e}\n")
            sys.stderr.flush()

def fetch(session, postcode):
    u = f"http://localhost:8000/postcode-search?postcode={postcode}"
    resp = session.get(u)
    if resp.status_code == 200:
        return resp.json()
    raise Exception("invalid status_code")



if __name__ == "__main__":
    main()
