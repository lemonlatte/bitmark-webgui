// Copyright (c) 2014-2016 Bitmark Inc.
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

'use strict';

/**
 * @ngdoc overview
 * @name bitmarkWebguiApp
 * @description
 * # bitmarkWebguiApp
 *
 * Main module of the application.
 */
var app = angular
  .module('bitmarkWebguiApp', [
    'ngCookies',
    'ngResource',
    'ngRoute',
    'ui.bootstrap'
  ])
        .constant("BitmarkProxyURL", {
            "testing": "https://spoon.test.bitmark.com:17555/rpc-call",
            "bitmark": "https://spoon.live.bitmark.com:17555/rpc-call"
        })
        .constant("ProxyTemp", {
            Username: "No-need-username",
            Password: "No-need-password",
            URL: "",
            Fee: "0.0002",
            Address: ""
        })
        .constant("BitmarkdConfig", {
            "testing": "/etc/bitmarkd-TESTING.conf",
            "bitmark": "/etc/bitmarkd-BITMARK.conf",
            "local": "/etc/bitmarkd-LOCAL.conf"
        })
        .constant("BitmarkCliSetupConfig", {
            identity: "admin",
            connect: "127.0.0.1:2130",
            description: "bitmark-webgui generated"
        })
        .constant("BitmarkCliConfig", {
            "testing": "/home/bitmark/config/bitmark-cli/bitmark-cli-TESTING.conf",
            "bitmark": "/home/bitmark/config/bitmark-cli/bitmark-cli-BITMARK.conf",
            "local": "/home/bitmark/config/bitmark-cli/bitmark-cli-LOCAL.conf"

            // "testing": "/home/yuntai/testWebgui/config/bitmark-cli/bitmark-cli-testing.config"
            // "local": "/home/yuntai/testWebgui/config/bitmark-cli/bitmark-cli-local.config"
        })
        .constant("BitmarkPayConfig", {
            "testing": "/home/bitmark/config/bitmark-pay/bitmark-pay-TESTING.xml",
            "bitmark": "/home/bitmark/config/bitmark-pay/bitmark-pay-BITMARK.xml",
            "local": "/home/bitmark/config/bitmark-pay/bitmark-pay-LOCAL.xml"

            // "testing": "/home/yuntai/testWebgui/config/bitmark-pay/bitmark-pay-TESTING.xml"
            // "local": "/home/yuntai/testWebgui/config/bitmark-pay/bitmark-pay-LOCAL.xml"
        })
        .config(function ($routeProvider, $httpProvider) {
            $httpProvider.defaults.withCredentials = true;
            delete $httpProvider.defaults.headers.common["X-Requested-With"];

            $routeProvider
                .when('/main', {
                    templateUrl: 'views/main.html',
                    controller: 'MainCtrl'
                })
                .when('/edit', {
                    templateUrl: 'views/edit.html',
                    controller: 'EditCtrl'
                })
                .when('/issue', {
                    templateUrl: 'views/issue.html',
                    controller: 'IssueCtrl'
                })
                .when('/transfer', {
                    templateUrl: 'views/transfer.html',
                    controller: 'TransferCtrl'
                })
                .when('/network', {
                    templateUrl: 'views/network.html',
                    controller: 'NetworkCtrl'
                })
                .when('/', {
                    templateUrl: 'views/login.html',
                    controller: 'LoginCtrl'
                })
                .when('/login', {
                    templateUrl: 'views/login.html',
                    controller: 'LoginCtrl'
                })
                .when('/login/create', {
                    templateUrl: 'views/loginCreate.html',
                    controller: 'LoginProcessCtrl'
                })
                .when('/login/access', {
                    templateUrl: 'views/loginAccess.html',
                    controller: 'LoginProcessCtrl'
                })
                .when('/logout', {
                    templateUrl: 'views/logout.html',
                    controller: 'LogoutCtrl'
                })
                .otherwise({
                    redirectTo: '/login'
                });
        });
