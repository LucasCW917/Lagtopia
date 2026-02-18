import bcrypt
import sqlite3
import socket, struct

def login_user(username: str, password: str, db_path: str) -> bool:
    if not username or not password:
        raise ValueError("username and password are required")

    if isinstance(password, str):
        password = password.encode('utf-8')

    with sqlite3.connect(db_path) as conn:
        cursor = conn.cursor()
        cursor.execute(
            "SELECT password FROM users WHERE username = ?",
            (username,),
        )
        row = cursor.fetchone()
        if row is None:
            return False

        stored_hash = row[0].encode('utf-8')
        return bcrypt.checkpw(password, stored_hash)

def send_message(sock, msg):
    data = msg.encode()
    length = struct.pack('>I', len(data))
    sock.sendall(length + data)

def recv_message(sock):
    length = struct.unpack('>I', sock.recv(4))[0]
    return sock.recv(length).decode()

sock = socket.create_connection(('localhost', 9000))
send_message(sock, "REGISTER LOGIN")

running = True
while running:
    msg = recv_message(sock)
    if msg.startswith("LOGIN"):
        _, username, password = msg.split() # syntax: LOGIN username password
        success = login_user(username, password, 'users.db')
        send_message(sock, "LOGIN_SUCCESS" if success else "LOGIN_FAILURE")
    elif msg == "QUIT":
        running = False