var myipApp = angular.module('myip', ['ui.bootstrap']);


myipApp.controller('MyIPController', function MyIPController($scope, $http) {
    MyIPController.$inject = ['$scope', '$http'];

    $scope.addresses = [];

    Object.keys(SERVERS).forEach(function(family) {
        var server = SERVERS[family];
        var url = server + '/json?family=' + family;
        $http.get(url).then(function success(response) {
            $scope.addresses.push(response.data);

        }, function error(response) {
            errorText = response["statusText"] || "unknown error";
            $scope.addresses.push({
                "RemoteAddrFamily": family,
                "Error": response["status"] + ": " + errorText
            });
        });
    });
});

myipApp.filter('firstWord', function firstWord($filter) {
    firstWord.$inject = ['$filter'];

    return function(data) {
        if(!data) return data;
        data = data.split(' ');
        return data[0];
    };
});

myipApp.filter('mapUrl', function mapUrl($filter) {
    mapUrl.$inject = ['$filter'];

    return function(data) {
        if (data) {
            var url = "https://maps.googleapis.com/maps/api/staticmap";
            url += "?key=" + MAPS_API_KEY;
            url += "&size=640x400";
            url += "&markers=color:red%7C";

            if ('Lat' in data && 'Long' in data && data['Lat'] != 0 && data['Long'] != 0) {
                return url + data['Lat'] + "," + data['Long']
            }

            if ('City' in data && data['City'] != "") {
                return url + data['City']
            }

            if ('Region' in data && data['Region'] != "") {
                return url + data['Region']
            }

            if ('Country' in data && data['Country'] != "") {
                return url + data['Country']
            }
        }

        return "";
    };
});

