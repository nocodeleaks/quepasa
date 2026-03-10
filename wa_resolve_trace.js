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

function readUtf16(ptr) {
  try {
    if (isNullPtr(ptr)) return null;
    return Memory.readUtf16String(ptr);
  } catch (e) {
    return null;
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

function sendEvent(kind, obj) {
  obj.kind = kind;
  obj.pid = Process.id;
  send(obj);
}

var gaiw = getExport('ws2_32.dll', 'GetAddrInfoW');
if (gaiw) {
  Interceptor.attach(gaiw, {
    onEnter: function (args) {
      this.node = readUtf16(args[0]);
      this.service = readUtf16(args[1]);
    },
    onLeave: function (retval) {
      sendEvent('GetAddrInfoW', {
        rc: retval.toInt32 ? retval.toInt32() : parseInt(retval.toString(), 10) || 0,
        node: this.node,
        service: this.service
      });
    }
  });
}

var gaiExw = getExport('ws2_32.dll', 'GetAddrInfoExW');
if (gaiExw) {
  Interceptor.attach(gaiExw, {
    onEnter: function (args) {
      this.node = readUtf16(args[0]);
      this.service = readUtf16(args[1]);
    },
    onLeave: function (retval) {
      sendEvent('GetAddrInfoExW', {
        rc: retval.toInt32 ? retval.toInt32() : parseInt(retval.toString(), 10) || 0,
        node: this.node,
        service: this.service
      });
    }
  });
}

var wcbn = getExport('ws2_32.dll', 'WSAConnectByNameW');
if (wcbn) {
  Interceptor.attach(wcbn, {
    onEnter: function (args) {
      this.sock = args[0].toString();
      this.node = readUtf16(args[1]);
      this.service = readUtf16(args[2]);
    },
    onLeave: function (retval) {
      sendEvent('WSAConnectByNameW', {
        sock: this.sock,
        rc: retval.toInt32 ? retval.toInt32() : parseInt(retval.toString(), 10) || 0,
        node: this.node,
        service: this.service
      });
    }
  });
}

var connectPtr = getExport('ws2_32.dll', 'connect');
if (connectPtr) {
  Interceptor.attach(connectPtr, {
    onEnter: function (args) {
      this.sock = args[0].toString();
      this.addr = readSockaddr(args[1]);
    },
    onLeave: function (retval) {
      sendEvent('connect', {
        sock: this.sock,
        rc: retval.toInt32 ? retval.toInt32() : parseInt(retval.toString(), 10) || 0,
        addr: this.addr
      });
    }
  });
}

var wsaconnectPtr = getExport('ws2_32.dll', 'WSAConnect');
if (wsaconnectPtr) {
  Interceptor.attach(wsaconnectPtr, {
    onEnter: function (args) {
      this.sock = args[0].toString();
      this.addr = readSockaddr(args[1]);
    },
    onLeave: function (retval) {
      sendEvent('WSAConnect', {
        sock: this.sock,
        rc: retval.toInt32 ? retval.toInt32() : parseInt(retval.toString(), 10) || 0,
        addr: this.addr
      });
    }
  });
}

console.log('resolver/connect tracer loaded pid=' + Process.id);
