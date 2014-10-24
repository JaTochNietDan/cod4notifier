$(document).ready(function() {
	var l = $('#btnRefresh').ladda();

	var timer;
	var lastServers = [];

	var ViewModel = function(servers) {
	    this.servers = ko.observableArray(servers);
	    this.currentServer = ko.observable({});
		this.notifyMe = ko.observable(false);
		this.selectedReload = ko.observable(0);

		if(support_storage()) {
			this.notifyMe(localStorage['notifyMe']);
		}

	    this.reloadEvery = ko.observableArray([
	    	{Time: 0, Name: 'Never'},
	    	{Time: 5, Name: '5 Seconds'},
	    	{Time: 10, Name: '10 Seconds'},
	    	{Time: 60, Name: 'Minute'},
	    	{Time: 5, Name: '5 Minutes'}
	    ]);

		this.selectedReload.subscribe(function(val) {
			if(val.Time != 0) {
				timer = setTimeout(GetServers, val.Time * 1000);
			} else {
				if(timer != 'undefined') {
					clearTimeout(timer);
				}
			}
		});

		this.notifyMe.subscribe(function(val) {
			if(support_storage()) {
				localStorage['notifyMe'] = val;
			}
		});
	};
	 
	var vm = new ViewModel([]);

	window.xxx = vm;

	ViewModel.prototype.ShowServer = function(server) {
		vm.currentServer(server);

		$('#modalServer').modal('show');
	};

	ko.applyBindings(vm); // This makes Knockout get to work

	$('#btnRefresh').click(function() {
		GetServers();
	});

	function GetServers() {
		l.ladda('start');

		$.getJSON('/servers.json', function(data) {
			if(data != null) {
				vm.servers(data);

				if(vm.notifyMe())
					checkServers();

				lastServers = data;
			}

			l.ladda('stop');

			if(vm.selectedReload().Time != 0) {
				timer = setTimeout(GetServers, vm.selectedReload().Time * 1000);
			}
		});
	}

	function checkServers() {
		vm.servers().forEach(function(server, i) {
			var found = false;

			// Look through last servers for it
			lastServers.forEach(function(old, x) {
				// Continue if it's the same hostname and same IP
				if(old.Hostname == server.Hostname && old.IP == server.IP) {
					found = true;
					return;
				}
			});

			if(!found)
				notifyNewServer(server);
		});
	}

	function notifyNewServer(server) {
		setupNotify();

		var notification = new Notification(server.Hostname, {
			icon: '/images/icon.png',
			body: 
				'Map: ' + server.Map + "\n" +
				'Gametype: ' + server.GameType + "\n",
		});


		notification.onclick = function () {
			vm.ShowServer(server);
		}
	}

	function support_storage() {
		try {
			return 'localStorage' in window && window['localStorage'] !== null;
		} catch (e) {
			return false;
		}
	} 
});

function setupNotify() {
	if (!Notification) {
		alert('Please us a modern version of Chrome, Firefox, Opera or Firefox.');
		return;
	}

	if (Notification.permission !== "granted")
		Notification.requestPermission();
}