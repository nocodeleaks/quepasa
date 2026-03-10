'use strict';

function getExport(moduleName, exportName) {
  try {
    return Process.getModuleByName(moduleName).getExportByName(exportName);
  } catch (e) {
    return null;
  }
}

function isNullPtr(ptr) {
  try {
    return ptr === null || ptr === undefined || ptr.toString() === '0x0';
  } catch (e) {
    return true;
  }
}

function ntohs(v) {
  return ((v & 0xff) << 8) | ((v >> 8) & 0xff);
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
      return {
        family: family,
        family_name: 'AF_INET',
        port: ntohs(Memory.readU16(ptr.add(2))),
        ip: [Memory.readU8(ptr.add(4)), Memory.readU8(ptr.add(5)), Memory.readU8(ptr.add(6)), Memory.readU8(ptr.add(7))].join('.')
      };
    }
    if (family === 23) {
      return {
        family: family,
        family_name: 'AF_INET6',
        port: ntohs(Memory.readU16(ptr.add(2))),
        ip: readIPv6(ptr.add(8)),
        flowinfo: Memory.readU32(ptr.add(4)) >>> 0,
        scope_id: Memory.readU32(ptr.add(24)) >>> 0
      };
    }
    return { family: family, family_name: 'AF_' + family };
  } catch (e) {
    return null;
  }
}

var getsocknamePtr = getExport('ws2_32.dll', 'getsockname');
var getpeernamePtr = getExport('ws2_32.dll', 'getpeername');
var getsocknameNative = getsocknamePtr ? new NativeFunction(getsocknamePtr, 'int', ['pointer', 'pointer', 'pointer']) : null;
var getpeernameNative = getpeernamePtr ? new NativeFunction(getpeernamePtr, 'int', ['pointer', 'pointer', 'pointer']) : null;

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

function sendEvent(kind, obj) {
  obj.kind = kind;
  obj.pid = Process.id;
  send(obj);
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

var connectExGuidHex = '2509e003dd0100000000000000000000';
var disconnectExGuidHex = '2709e003dd0100000000000000000000';
var extensionHooks = {};

function guidBytesToHex(ptr) {
  try {
    var s = '';
    for (var i = 0; i < 16; i++) {
      var h = Memory.readU8(ptr.add(i)).toString(16);
      if (h.length < 2) h = '0' + h;
      s += h;
    }
    return s;
  } catch (e) {
    return null;
  }
}

function hookConnectEx(fnPtr) {
  var key = fnPtr.toString();
  if (extensionHooks[key]) return;
  extensionHooks[key] = true;
  Interceptor.attach(fnPtr, {
    onEnter: function (args) {
      this.sock = args[0];
      this.name = readSockaddr(args[1]);
      this.namelen = safeLen(args[2]);
      this.sendlen = safeLen(args[4]);
    },
    onLeave: function (retval) {
      sendEvent('ConnectEx', {
        sock: this.sock.toString(),
        rc: safeLen(retval),
        addr: this.name,
        namelen: this.namelen,
        sendlen: this.sendlen,
        local: querySockAddr(this.sock, 'local'),
        remote: querySockAddr(this.sock, 'peer')
      });
    }
  });
  sendEvent('hook', { name: 'ConnectEx', ptr: fnPtr.toString() });
}

var wsaiocltPtr = getExport('ws2_32.dll', 'WSAIoctl');
if (wsaiocltPtr) {
  Interceptor.attach(wsaiocltPtr, {
    onEnter: function (args) {
      this.sock = args[0];
      this.code = args[1].toString();
      this.inbuf = args[2];
      this.inlen = safeLen(args[3]);
      this.outbuf = args[4];
      this.outlen = safeLen(args[5]);
      this.guidHex = this.inlen >= 16 ? guidBytesToHex(this.inbuf) : null;
    },
    onLeave: function (retval) {
      var rc = safeLen(retval);
      if (this.guidHex === connectExGuidHex && this.outlen >= Process.pointerSize) {
        try {
          var fnPtr = Memory.readPointer(this.outbuf);
          sendEvent('WSAIoctl', {
            sock: this.sock.toString(),
            rc: rc,
            code: this.code,
            guid_hex: this.guidHex,
            fn_ptr: fnPtr.toString()
          });
          hookConnectEx(fnPtr);
          return;
        } catch (e) {
        }
      }
      if (this.guidHex === disconnectExGuidHex) {
        sendEvent('WSAIoctl', {
          sock: this.sock.toString(),
          rc: rc,
          code: this.code,
          guid_hex: this.guidHex,
          note: 'DisconnectEx GUID requested'
        });
      }
    }
  });
}

var connectPtr = getExport('ws2_32.dll', 'connect');
if (connectPtr) {
  Interceptor.attach(connectPtr, {
    onEnter: function (args) {
      this.sock = args[0];
      this.addr = readSockaddr(args[1]);
    },
    onLeave: function (retval) {
      sendEvent('connect', {
        sock: this.sock.toString(),
        rc: safeLen(retval),
        addr: this.addr,
        local: querySockAddr(this.sock, 'local'),
        remote: querySockAddr(this.sock, 'peer')
      });
    }
  });
}

var wsaconnectPtr = getExport('ws2_32.dll', 'WSAConnect');
if (wsaconnectPtr) {
  Interceptor.attach(wsaconnectPtr, {
    onEnter: function (args) {
      this.sock = args[0];
      this.addr = readSockaddr(args[1]);
    },
    onLeave: function (retval) {
      sendEvent('WSAConnect', {
        sock: this.sock.toString(),
        rc: safeLen(retval),
        addr: this.addr,
        local: querySockAddr(this.sock, 'local'),
        remote: querySockAddr(this.sock, 'peer')
      });
    }
  });
}

console.log('connectex tracer loaded pid=' + Process.id);
