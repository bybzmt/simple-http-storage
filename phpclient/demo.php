<?php
namespace Bybzmt\HttpStorage;

require "Exception.php";
require "SimpleHttpStorage.php";

$server = "127.0.0.1";
$port = 8080;
$timeout = 10;

$storage = new SimpleHttpStorage($server, $port, $timeout);

$storage->put(__FILE__, '/test.txt');

$storage->put(__FILE__, '/mkdir/123/test.txt');

var_dump($storage->exists("/test.txt"));

$storage->get("/test.txt", "./test.txt");

$storage->delete("/test.txt");
