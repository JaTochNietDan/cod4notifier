$(document).ready(function() {
	var l = $('#btnRefresh').ladda();

	var ViewModel = function(servers) {
	    this.servers = ko.observableArray(servers);
	    this.currentServer = ko.observable({});
	    this.notifyMe = ko.observable(false);

	    this.reloadEvery = ko.observableArray([
	    	{Time: 0, Name: 'Never'},
	    	{Time: 1, Name: 'Second'},
	    	{Time: 10, Name: '10 Seconds'},
	    	{Time: 60, Name: 'Minute'},
	    	{Time: 5, Name: '5 Minutes'}
	    ]);
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
			vm.servers(data);

			l.ladda('stop');
		});
	}
});