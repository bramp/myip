// Copyright 2017 Google Inc. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

var myipApp = angular.module('myip', ['ui.bootstrap'], function MyIPApp($locationProvider) {
    MyIPApp.$inject = ['$locationProvider'];

    $locationProvider.html5Mode(true).hashPrefix('');
});

myipApp.controller('MyIPController', function MyIPController($scope, $http, $location) {
    MyIPController.$inject = ['$scope', '$http', '$location'];

    $scope.version = VERSION;
    $scope.buildTime = BUILDTIME;

    var host = $location.search().host;
    $scope.addresses = [];

    Object.keys(SERVERS).forEach(function(family) {
        var server = SERVERS[family];
        var url = $location.protocol() + "://" + server + '/json?family=' + family;

        if (host) {
            url = url + "&host=" + host
        }

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
