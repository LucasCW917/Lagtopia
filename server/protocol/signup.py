import os
import bcrypt
import sqlite3
from typing import Optional

BASE_DIR = os.path.abspath(os.path.join(os.path.dirname(__file__), '..'))
DB_PATH = os.path.join(BASE_DIR, 'users.db')


def init_db(conn: sqlite3.Connection) -> None:
    cursor = conn.cursor()
    cursor.execute(
        """
        CREATE TABLE IF NOT EXISTS users (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            username TEXT UNIQUE NOT NULL,
            password TEXT NOT NULL
        )
        """
    )
    conn.commit()


def signup_user(username: str, password: str, db_path: Optional[str] = None) -> bool:
    if not username or not password:
        raise ValueError("username and password are required")

    if isinstance(password, str):
        password = password.encode('utf-8')

    hashed = bcrypt.hashpw(password, bcrypt.gensalt())
    hashed_str = hashed.decode('utf-8')

    db = db_path or DB_PATH
    os.makedirs(os.path.dirname(db), exist_ok=True)
    with sqlite3.connect(db) as conn:
        init_db(conn)
        cursor = conn.cursor()
        try:
            cursor.execute(
                "INSERT INTO users (username, password) VALUES (?, ?)",
                (username, hashed_str),
            )
            conn.commit()
            return True
        except sqlite3.IntegrityError:
            return False