import socket
import struct

class TCPClient:
    def __init__(self, host="localhost", port=9000):
        self.sock = socket.create_connection((host, port))

    def send(self, message: str):
        data = message.encode()
        length = struct.pack(">I", len(data))  # 4-byte big endian
        self.sock.sendall(length + data)

    def receive(self):
        length_bytes = self._recv_exact(4)
        length = struct.unpack(">I", length_bytes)[0]
        data = self._recv_exact(length)
        return data.decode()

    def _recv_exact(self, n):
        data = b""
        while len(data) < n:
            chunk = self.sock.recv(n - len(data))
            if not chunk:
                raise ConnectionError("Connection closed")
            data += chunk
        return data
