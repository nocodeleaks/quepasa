'use strict';

var tracked = {};

function ntohs(v) {
  return ((v & 0xff) << 8) | ((v >> 8) & 0xff);
}

function getExport(moduleName, exportName) {
  try {
    return Process.getModuleByName(moduleName).getExportByName(exportName);
  } catch (e) {
    return null;
  }
}

var getsocknamePtr = getExport('ws2_32.dll', 'getsockname');
var getpeernamePtr = getExport('ws2_32.dll', 'getpeername');
var getsocknameNative = getsocknamePtr ? new NativeFunction(getsocknamePtr, 'int', ['pointer', 'pointer', 'pointer']) : null;
var getpeernameNative = getpeernamePtr ? new NativeFunction(getpeernamePtr, 'int', ['pointer', 'pointer', 'pointer']) : null;

function isNullPtr(ptr) {
  try {
    return ptr === null || ptr === undefined || ptr.toString() === '0x0';
  } catch (e) {
    return true;
  }
}

function readIPv6(ptr) {
  var parts = [];
  for (var i = 0; i < 16; i += 2) {
    var hi = Memory.readU8(ptr.add(i));
    var lo = Memory.readU8(ptr.add(i + 1));
    parts.push(((hi << 8) | lo).toString(16));
  }
  return parts.join(':');
}

function readSockaddr(ptr) {
  try {
    if (isNullPtr(ptr)) return null;
    var family = Memory.readU16(ptr);
    if (family === 2) {
      var portNet4 = Memory.readU16(ptr.add(2));
      var port4 = ntohs(portNet4);
      var b0 = Memory.readU8(ptr.add(4));
      var b1 = Memory.readU8(ptr.add(5));
      var b2 = Memory.readU8(ptr.add(6));
      var b3 = Memory.readU8(ptr.add(7));
      return { family: family, family_name: 'AF_INET', port: port4, ip: b0 + '.' + b1 + '.' + b2 + '.' + b3 };
    }
    if (family === 23) {
      var portNet6 = Memory.readU16(ptr.add(2));
      var port6 = ntohs(portNet6);
      var flowinfo = Memory.readU32(ptr.add(4));
      var ip6 = readIPv6(ptr.add(8));
      var scope = Memory.readU32(ptr.add(24));
      return { family: family, family_name: 'AF_INET6', port: port6, ip: ip6, flowinfo: flowinfo >>> 0, scope_id: scope >>> 0 };
    }
    return { family: family, family_name: 'AF_' + family };
  } catch (e) {
    return null;
  }
}

function querySockAddr(sock, which) {
  try {
    var fn = which === 'peer' ? getpeernameNative : getsocknameNative;
    if (!fn) return null;
    var buf = Memory.alloc(64);
    var lenPtr = Memory.alloc(4);
    Memory.writeU32(lenPtr, 64);
    var rc = fn(sock, buf, lenPtr);
    if (rc !== 0) return null;
    return readSockaddr(buf);
  } catch (e) {
    return null;
  }
}

function readBytes(ptr, len, maxLen) {
  var out = [];
  try {
    if (isNullPtr(ptr)) return out;
    var n = len;
    if (n < 0) n = 0;
    if (maxLen && n > maxLen) n = maxLen;
    for (var i = 0; i < n; i++) out.push(Memory.readU8(ptr.add(i)));
  } catch (e) {
    return [];
  }
  return out;
}

function isStunBytes(bytes) {
  if (!bytes || bytes.length < 20) return false;
  if ((bytes[0] & 0xc0) !== 0x00) return false;
  var cookie = ((bytes[4] << 24) >>> 0) | (bytes[5] << 16) | (bytes[6] << 8) | bytes[7];
  return cookie === 0x2112A442;
}

function toHex(bytes) {
  var s = '';
  for (var i = 0; i < bytes.length; i++) {
    var h = bytes[i].toString(16);
    if (h.length < 2) h = '0' + h;
    s += h;
  }
  return s;
}

function safeLen(v) {
  try {
    return v.toInt32();
  } catch (e) {
    try {
      return parseInt(v.toString(), 10) || 0;
    } catch (e2) {
      return 0;
    }
  }
}

function sockKey(sock) {
  try {
    return sock.toString();
  } catch (e) {
    return 'unknown';
  }
}

function ensureSock(sock) {
  var key = sockKey(sock);
  if (!tracked[key]) tracked[key] = { sock: key };
  return tracked[key];
}

function enrichMeta(sock, meta) {
  if (!meta.local) meta.local = querySockAddr(sock, 'local');
  if (!meta.remote) meta.remote = querySockAddr(sock, 'peer');
}

function sendEvent(kind, obj) {
  obj.kind = kind;
  obj.pid = Process.id;
  send(obj);
}

function maybeLogData(api, dir, sock, addr, buf, len) {
  var meta = ensureSock(sock);
  enrichMeta(sock, meta);
  var bytes = readBytes(buf, len, 64);
  if (bytes.length === 0) return;
  var stun = isStunBytes(bytes);
  var interesting = stun || (len > 0 && len <= 512);
  if (!interesting) return;
  sendEvent('data', {
    api: api,
    dir: dir,
    sock: sockKey(sock),
    addr: addr,
    local: meta.local || null,
    remote: meta.remote || null,
    len: len,
    stun: stun,
    hex: toHex(bytes)
  });
}

function attachSocketStyle(name, exportName, afIndex, typeIndex, protoIndex) {
  var ptr = getExport('ws2_32.dll', exportName);
  if (!ptr) return;
  Interceptor.attach(ptr, {
    onEnter: function (args) {
      this.af = safeLen(args[afIndex]);
      this.type = safeLen(args[typeIndex]);
      this.proto = safeLen(args[protoIndex]);
    },
    onLeave: function (retval) {
      var sock = retval;
      var meta = ensureSock(sock);
      meta.af = this.af;
      meta.type = this.type;
      meta.proto = this.proto;
      sendEvent(name, { sock: sockKey(sock), af: this.af, type: this.type, proto: this.proto });
    }
  });
}

function attachBindLike(name, exportName, sockIndex, addrIndex) {
  var ptr = getExport('ws2_32.dll', exportName);
  if (!ptr) return;
  Interceptor.attach(ptr, {
    onEnter: function (args) {
      this.sock = args[sockIndex];
      this.addr = readSockaddr(args[addrIndex]);
    },
    onLeave: function (retval) {
      var meta = ensureSock(this.sock);
      meta.local = this.addr || querySockAddr(this.sock, 'local');
      sendEvent(name, { sock: sockKey(this.sock), rc: safeLen(retval), addr: meta.local || this.addr || null });
    }
  });
}

function attachConnectLike(name, exportName, sockIndex, addrIndex) {
  var ptr = getExport('ws2_32.dll', exportName);
  if (!ptr) return;
  Interceptor.attach(ptr, {
    onEnter: function (args) {
      this.sock = args[sockIndex];
      this.addr = readSockaddr(args[addrIndex]);
    },
    onLeave: function (retval) {
      var meta = ensureSock(this.sock);
      meta.remote = this.addr || querySockAddr(this.sock, 'peer');
      meta.local = meta.local || querySockAddr(this.sock, 'local');
      sendEvent(name, { sock: sockKey(this.sock), rc: safeLen(retval), addr: meta.remote || this.addr || null, local: meta.local || null });
    }
  });
}

function attachClose() {
  var ptr = getExport('ws2_32.dll', 'closesocket');
  if (!ptr) return;
  Interceptor.attach(ptr, {
    onEnter: function (args) {
      this.sock = args[0];
      var meta = ensureSock(this.sock);
      enrichMeta(this.sock, meta);
      sendEvent('close', { sock: sockKey(this.sock), meta: meta });
      delete tracked[sockKey(this.sock)];
    }
  });
}

function attachSendtoLike(name, exportName, bufIndex, lenIndex, toIndex) {
  var ptr = getExport('ws2_32.dll', exportName);
  if (!ptr) return;
  Interceptor.attach(ptr, {
    onEnter: function (args) {
      try {
        var sock = args[0];
        var buf = args[bufIndex];
        var len = safeLen(args[lenIndex]);
        var addr = readSockaddr(args[toIndex]);
        maybeLogData(name, 'out', sock, addr, buf, len);
      } catch (e) {
      }
    }
  });
}

function attachWSABufLike(name, exportName, lpBuffersIndex, countIndex, toIndex) {
  var ptr = getExport('ws2_32.dll', exportName);
  if (!ptr) return;
  Interceptor.attach(ptr, {
    onEnter: function (args) {
      try {
        var sock = args[0];
        var lpBuffers = args[lpBuffersIndex];
        var count = safeLen(args[countIndex]);
        var addr = readSockaddr(args[toIndex]);
        for (var i = 0; i < count; i++) {
          var wsabuf = lpBuffers.add(i * Process.pointerSize * 2);
          var len = Memory.readU32(wsabuf);
          var buf = Memory.readPointer(wsabuf.add(Process.pointerSize));
          maybeLogData(name, 'out', sock, addr, buf, len);
        }
      } catch (e) {
      }
    }
  });
}

function attachRecvfromLike(name, exportName, bufIndex, fromIndex) {
  var ptr = getExport('ws2_32.dll', exportName);
  if (!ptr) return;
  Interceptor.attach(ptr, {
    onEnter: function (args) {
      this.sock = args[0];
      this.buf = args[bufIndex];
      this.from = args[fromIndex];
    },
    onLeave: function (retval) {
      try {
        var len = safeLen(retval);
        if (len <= 0) return;
        var addr = readSockaddr(this.from);
        maybeLogData(name, 'in', this.sock, addr, this.buf, len);
      } catch (e) {
      }
    }
  });
}

function attachWSARecvLike(name, exportName, lpBuffersIndex, countIndex) {
  var ptr = getExport('ws2_32.dll', exportName);
  if (!ptr) return;
  Interceptor.attach(ptr, {
    onEnter: function (args) {
      this.sock = args[0];
      this.lpBuffers = args[lpBuffersIndex];
      this.count = safeLen(args[countIndex]);
    },
    onLeave: function (retval) {
      try {
        if (safeLen(retval) < 0) return;
        for (var i = 0; i < this.count; i++) {
          var wsabuf = this.lpBuffers.add(i * Process.pointerSize * 2);
          var len = Memory.readU32(wsabuf);
          var buf = Memory.readPointer(wsabuf.add(Process.pointerSize));
          maybeLogData(name, 'in', this.sock, null, buf, len);
        }
      } catch (e) {
      }
    }
  });
}

attachSocketStyle('socket', 'socket', 0, 1, 2);
attachSocketStyle('WSASocketW', 'WSASocketW', 0, 1, 2);
attachSocketStyle('WSASocketA', 'WSASocketA', 0, 1, 2);
attachBindLike('bind', 'bind', 0, 1);
attachConnectLike('connect', 'connect', 0, 1);
attachConnectLike('WSAConnect', 'WSAConnect', 0, 1);
attachClose();
attachSendtoLike('sendto', 'sendto', 1, 2, 4);
attachSendtoLike('send', 'send', 1, 2, 3);
attachWSABufLike('WSASendTo', 'WSASendTo', 1, 2, 5);
attachWSABufLike('WSASend', 'WSASend', 1, 2, 4);
attachRecvfromLike('recvfrom', 'recvfrom', 1, 4);
attachRecvfromLike('recv', 'recv', 1, 3);
attachWSARecvLike('WSARecv', 'WSARecv', 1, 2);

console.log('socket tracer loaded pid=' + Process.id);
