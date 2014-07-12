var Turbo = (function() {
    'use strict';

    var MSG_CMD_ON = 1,
        MSG_CMD_OFF = 2,
        MSG_CMD_SET = 3,
        MSG_CMD_UPDATE = 4,
        MSG_CMD_REMOVE = 5,
        MSG_CMD_TRANS_SET = 6,
        MSG_CMD_PUSH = 7,
        MSG_CMD_TRANS_GET = 8,
        MSG_CMD_AUTH = 9,
        MSG_CMD_UNAUTH = 10,
        MSG_CMD_ACK = 11;

    var EVENT_TYPE_VALUE = 0,
        EVENT_TYPE_CHILD_ADDED = 1,
        EVENT_TYPE_CHILD_CHANGED = 2,
        EVENT_TYPE_CHILD_MOVED = 3,
        EVENT_TYPE_CHILD_REMOVED = 4;

    var EVENT_TYPE_VALUE_STR = 'value',
        EVENT_TYPE_CHILD_ADDED_STR = 'child_added',
        EVENT_TYPE_CHILD_CHANGED_STR = 'child_changed',
        EVENT_TYPE_CHILD_MOVED_STR = 'child_moved',
        EVENT_TYPE_CHILD_REMOVED_STR = 'child_removed';

    var _ws = undefined;
    var _listeners = {};
    var _ack = 0;
    var _ackCallbacks = {};
    var _token = undefined;
    var _isOffline = true;
    var _offlineQueue = [];

    var _send = function _send(val) {
        console.log("Sending: ", val);
        if (_isOffline) {
            _offlineQueue.push(val);
        } else {
            if (!_ws) throw 'Must be connected to the server to send a message';
            _ws.send(val);
        }
    };

    var _connect = function _connect(url, onConnect) {
        console.log('Connecting to ' + url);
        _ws = new WebSocket(url);
        _ws.onopen = function(evt) {
            _isOffline = false;
            if (onConnect) onConnect(evt);
            console.log('Connection opened.', evt);
        };
        _ws.onclose = function(evt) {
            _isOffline = true;
            console.log('Connection closed.', evt);
        };
        _ws.onerror = function(evt) {
            console.log('Connection had an error.', evt);
        };
        _ws.onmessage = function(evt) {
            if (!evt) return;
            if (!evt.data) return;

            try {
                var msg = JSON.parse(evt.data);
                switch (msg.type) {
                    case MSG_CMD_ACK:
                        if (_ackCallbacks[msg.ack]) {
                            _ackCallbacks[msg.ack](msg.err, msg.res, msg.hash);
                            delete _ackCallbacks[msg.ack];
                        }
                        break;
                    default:
                        if (!msg.eventType || !msg.path) return; // Filter for 'on' events

                        if (_listeners[msg.path] && _listeners[msg.path][msg.eventType]) {
                            var listenerMap = _listeners[msg.path][msg.eventType];
                            for (var listenerRef in listenerMap) {
                                var listener;
                                if (listener = listenerMap[listenerRef]) {
                                    var context = listener.context || listenerRef;
                                    if (listener.callback) {
                                        listener.callback.call(context, new DataSnapShot(msg.value, this._url, msg.path));
                                    }
                                }
                            }
                        }
                }
            } catch (e) {
                console.log('Could not process message; json parsing is failing', e);
            }
        };
    };

    var _disconnect = function _disconnect() {
        console.log('Disconnecting');
        _isOffline = true;
        _ws.close();
        _ws = undefined;
    };

    var _attemptTransSet = function _attemptTransSet(path, value, hash, transform, done) {
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': MSG_CMD_TRANS_SET,
            'path': path,
            'hash': hash,
            'value': transform(value),
            'ack': ack
        }));
        _ackCallbacks[ack] = function(err, newValue, newHash) {
            if (err === 'conflict') _attemptTransSet(path, newValue, newHash, transform, done);
            else if (err) done(err);
            else done(undefined, newValue);
        };
    };

    var _eventType = function _eventType(eventTypeStr) {
        switch (eventTypeStr) {
            case EVENT_TYPE_VALUE_STR:
                return EVENT_TYPE_VALUE;
            case EVENT_TYPE_CHILD_ADDED_STR:
                return EVENT_TYPE_CHILD_ADDED;
            case EVENT_TYPE_CHILD_CHANGED_STR:
                return EVENT_TYPE_CHILD_CHANGED;
            case EVENT_TYPE_CHILD_MOVED_STR:
                return EVENT_TYPE_CHILD_MOVED;
            case EVENT_TYPE_CHILD_REMOVED:
                return EVENT_TYPE_CHILD_REMOVED;
        }
        return null;
    };

    var Client = function(url, path) {
        if (url === null || url === undefined || !(typeof url === 'string'))
            throw new Error('url was invalid');
        if (path === null || path === undefined || !(typeof path === 'string'))
            path = '/';

        if (url.charAt(0) === '/') url = window.location.host + url;
        if (url.indexOf('ws://') !== 0) url = 'ws://' + url;
        if (url.charAt(url.length - 1) === '/') url = url.slice(0, -1);

        this._url = url;
        this._path = path;

        if (!_ws) _connect(url, function(evt) {
            var msg;
            while ((msg = _offlineQueue.shift())) {
                _ws.send(msg);
            }
        });
    };

    Client.prototype.on = function(eventTypeStr, callback, cancelCallback, context) {
        var eventType = _eventType(eventTypeStr);
        if (!eventType)
            throw 'Unsupported event type \'' + eventTypeStr + '\'';

        if (!callback || typeof callback !== 'function') throw 'Callback was not a function';
        if (cancelCallback && !context && typeof cancelCallback === 'object') {
            context = cancelCallback;
            cancelCallback = undefined;
        }

        var self = this;
        var path = self._path;
        _send(JSON.stringify({
            'cmd': MSG_CMD_ON,
            'eventType': eventType,
            'path': path
        }));

        if (!_listeners[path]) _listeners[path] = {};
        if (!_listeners[path][eventType]) _listeners[path][eventType] = {};
        _listeners[path][eventType][self] = {
            callback: callback,
            context: context,
            cancelCallback: cancelCallback
        };

        return callback;
    };

    Client.prototype.off = function(eventType, callback, context) {
        var eventType = _eventType(eventTypeStr);
        if (!eventType)
            throw 'Unsupported event type \'' + eventTypeStr + '\'';

        if (!callback || typeof callback !== 'function') throw 'Callback was not a function';

        var self = this;
        var path = self._path;

        if (!_listeners[path]) return;
        if (!_listeners[path][eventType]) return;
        if (!_listeners[path][eventType][self]) return;
        delete _listeners[path][eventType][self];

        _send(JSON.stringify({
            'cmd': MSG_CMD_OFF,
            'eventType': eventType,
            'path': path
        }));
    };

    Client.prototype.child = function(childPath) {
        if (!childPath) return this;
        if (childPath.length >= 1 && childPath[0] === '/') childPath = childPath.substring(1)
        var newPath = this._path === '/' ? ('/' + childPath) : (this._path + '/' + childPath);
        return new Client(this._url, newPath);
    };

    Client.prototype.parent = function() {
        if (this._path === '/') return this;
        var newPath = this._path.slice(0, this._path.lastIndexOf('/'));
        return new Client(this._url, newPath);
    };

    Client.prototype.root = function() {
        if (this._path === '/') return this;
        return new Client(this._url, '/');
    };

    Client.prototype.toString = function() {
        return this._path;
    };

    Client.prototype.set = function(value, onComplete) {
        var self = this;
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': MSG_CMD_SET,
            'path': self._path,
            'value': value,
            'ack': ack
        }));
        _ackCallbacks[ack] = onComplete;
    };

    Client.prototype.update = function(value, onComplete) {
        var self = this;
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': MSG_CMD_UPDATE,
            'path': self._path,
            'value': value,
            'ack': ack
        }));
        _ackCallbacks[ack] = onComplete;
    };

    Client.prototype.name = function() {
        return this._path.split('/').pop();
    };

    Client.prototype.setWithPriority = function(newVal, newPriority, onComplete) {
        throw 'Turbo does not support setWithPriority(...) right now';
    };

    Client.prototype.remove = function(onComplete) {
        var self = this;
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': MSG_CMD_REMOVE,
            'path': self._path,
            'ack': ack
        }));
        _ackCallbacks[ack] = onComplete;
    };

    Client.prototype.transaction = function(transactionUpdate, onComplete, applyLocally) {
        var self = this;
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': MSG_CMD_TRANS_GET,
            'path': self._path,
            'ack': ack
        }));
        _ackCallbacks[ack] = function(err, value, hash) {
            if (err) onComplete(err);
            else _attemptTransSet(self._path, value, hash, transactionUpdate, onComplete);
        };
    };

    Client.prototype.setPriority = function(priority, opt_onComplete) {
        throw 'Turbo does not support setPriority(...) right now';
    };

    Client.prototype.push = function(value, onComplete) {
        var self = this;
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': MSG_CMD_PUSH,
            'path': self._path,
            'value': value,
            'ack': ack
        }));
        _ackCallbacks[ack] = onComplete;
    };

    Client.prototype.onDisconnect = function() {
        throw 'Turbo does not support onDisconnect(...) right now';
    };

    Client.prototype.removeOnDisconnect = function() {
        throw 'Turbo does not support removeOnDisconnect(...) right now';
    };

    Client.prototype.setOnDisconnect = function(onDc) {
        throw 'Turbo does not support setOnDisconnect(...) right now';
    };

    Client.prototype.auth = function(cred, onComplete, onCancel) {
        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': MSG_CMD_AUTH,
            'cred': cred,
            'ack': ack
        }));
        _ackCallbacks[ack] = onComplete;
    };

    Client.prototype.unauth = function(onComplete) {
        if (!_token) throw 'Cannot unauth if not authed yet';

        var ack = _ack++;
        _send(JSON.stringify({
            'cmd': MSG_CMD_UNAUTH,
            'token': _token,
            'ack': ack
        }));
        _ackCallbacks[ack] = function(err, res) {
            _token = undefined;
            onComplete(err, res);
        };
    };

    Client.prototype.goOffline = function() {
        _disconnect();
    };

    Client.prototype.goOnline = function() {
        _connect(this._url, function(evt) {
            var msg;
            while ((msg = _offlineQueue.shift())) {
                _ws.send(msg);
            }
        });
    };

    Client.prototype.enableLogging = function(logger, persistent) {
        throw 'Turbo does not support enableLogging(...) right now';
    };
	
	function DataSnapshot(baseObj, url, path){
		this._baseObj = baseObj;
		this._url = url;
		this._path = path;
	}

	DataSnapshot.prototype.val = function(){
		return this._baseObj;
	};

	DataSnapshot.prototype.child = function(childName){
		return new DataSnapshot(this._baseObj[childName]);
	}

	DataSnapshot.prototype.forEach = function(childAction){
		for(var child in this._baseObj){
			childAction(child);
		}
	};

	DataSnapshot.prototype.hasChild = function(childName){
		return !!this._baseObj[childName];
	};

	DataSnapshot.prototype.hasChildren = function(){
		return Object.keys(this._baseObj).length != 0;
	};

	DataSnapshot.prototype.name = function(){
		return this._baseObj.split('/').pop();
	};

	DataSnapshot.prototype.numChildren = function(){
		return Object.keys(this._baseObj).length;
	};

	DataSnapshot.prototype.ref = function(){
		return new Client(this._url, this._path);
	};

	DataSnapshot.prototype.getPriority = function(){
		//TODO: are we doing this?
	};

	DataSnapshot.prototype.exportVal = function(){
		//TODO: are we doing priority? if not this is the same as val()
	};

    return Client;
})();