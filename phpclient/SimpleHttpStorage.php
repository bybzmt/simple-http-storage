<?php
namespace Bybzmt\HttpStorage;

/**
 * 简易HTTP文件系统PHP客户端
 * 这个是标准HTTP协议模式
 */
class SimpleHttpStorage
{
	private $server;
	private $port;
	private $timeout;
	private $head_cache;

	/**
	 * 设置配置
	 */
	public function __construct($server, $port, $timeout)
	{
		$this->server = $server;
		$this->port = $port;
		$this->timeout = $timeout;
	}

	/**
	 * 测试一个文件是否存在
	 */
	public function exists($remote_file)
	{
		$has = $this->_head($remote_file);
		return $has ? true : false;
	}

	/**
	 * 下载远程文件到本地
	 */
	public function get($remote_file, $local_file)
	{
		$dst = fopen($local_file, 'wb');
		if (!$dst) { return false; }

		$src = $this->_open($remote_file);
		if (!$src) {
			fclose($dst);
			return false;
		}

		stream_copy_to_stream($src, $dst);

		fclose($dst);
		fclose($src);

		return true;
	}

	/**
	 * 上传本地文件到远程
	 */
	public function put($local_file, $remote_file)
	{
		$src = fopen($local_file, 'rb');
		if (!$src) { return false; }

		$ok = $this->stream_put($src, $remote_file);
		fclose($src);

		return $ok;
	}

	/**
	 * 删除远程文件
	 */
	public function delete($remote_file)
	{
		$file = $this->_parseFile($remote_file);
		if ($file===false) { return false; }

		unset($this->head_cache[$file]);

		$req = array();
		$req[] = "DELETE /{$file} HTTP/1.0";
		$req[] = "Host: {$this->server}";
		$req[] = "Content-Length: 0";
		$req[] = "\r\n";
		$req = implode("\r\n", $req);

		$dst = fsockopen($this->server, $this->port, $errno, $error, $this->timeout);
		if (!$dst) {
			throw new Exception("SimpleHttpStorage Errno: {$errno} Error: {$error}");
		}

		fwrite($dst, $req);

		$status = stream_get_line($dst, 1024, "\r\n");
		$msg = stream_get_contents($dst, 4096);
		fclose($dst);

		if ($status != 'HTTP/1.0 200 OK') {
			trigger_error("DELETE {$remote_file} Fail {$msg}");
			return false;
		}

		return true;
	}

	/**
	 * 打开一个读取流
	 */
	public function stream_get($remote_file)
	{
		return $this->_open($remote_file);
	}

	/**
	 * 从流上传到文件
	 */
	public function stream_put($src, $remote_file)
	{
		if (!is_resource($src)) {
			return false;
		}
		return $this->_put($src, $remote_file);
	}

	/**
	 * 从数据上传到文件
	 */
	public function data_put($data, $file)
	{
		if (!is_string($data)) {
			return false;
		}
		return $this->_put($data, $file);
	}

	/**
	 * 打开连接
	 */
	private function _open($file)
	{
		$file = $this->_parseFile($file);
		if ($file===false) { return false; }

		$req = array();
		$req[] = "GET /{$file} HTTP/1.0";
		$req[] = "Host: {$this->server}";
		$req[] = "Content-Length: 0";
		$req[] = "\r\n";
		$req = implode("\r\n", $req);

		$dst = fsockopen($this->server, $this->port, $errno, $error, $this->timeout);
		if (!$dst) {
			throw new Exception("SimpleHttpStorage Errno: {$errno} Error: {$error}");
		}

		$ok = fwrite($dst, $req);
		if (!$ok) {
			return false;
		}

		list($code, $head, $dst) = $this->_getResponse($dst, "1.0", true);
		if ($code != '200') {
			fclose($dst);
			$this->head_cache[$file] = false;
			return false;
		}

		$this->head_cache[$file] = $head;

		return $dst;
	}

	/**
	 * 通HEAD请求取得头信息
	 */
	private function _head($file)
	{
		$file = $this->_parseFile($file);
		if ($file===false) { return false; }

		if (isset($this->head_cache[$file])) {
			return $this->head_cache[$file];
		}

		$req = array();
		$req[] = "HEAD /{$file} HTTP/1.0";
		$req[] = "Host: {$this->server}";
		$req[] = "Content-Length: 0";
		$req[] = "\r\n";
		$req = implode("\r\n", $req);

		$dst = fsockopen($this->server, $this->port, $errno, $error, $this->timeout);
		if (!$dst) {
			throw new Exception("SimpleHttpStorage Errno: {$errno} Error: {$error}");
		}

		$ok = fwrite($dst, $req);
		if (!$ok) {
			return false;
		}

		list($code, $head, $msg) = $this->_getResponse($dst, "1.0");
		if ($code != '200') {
			$this->head_cache[$file] = false;
			return false;
		}

		$this->head_cache[$file] = $head;

		return $head;
	}

	/**
	 * 上传文件
	 */
	private function _put($src, $file)
	{
		$file = $this->_parseFile($file);
		if ($file===false) { return false; }

		unset($this->head_cache[$file]);

		$req = array();
		$req[] = "PUT /{$file} HTTP/1.1";
		$req[] = "Host: {$this->server}";
		$req[] = "Connection: close";
		$req[] = "Transfer-Encoding: chunked";
		$req[] = "Content-Type: application/octet-stream";
		//head空行
		$req[] = "\r\n";
		$req = implode("\r\n", $req);

		$dst = fsockopen($this->server, $this->port, $errno, $error, $this->timeout);
		if (!$dst) {
			throw new Exception("SimpleHttpStorage Errno: {$errno} Error: {$error}");
		}

		//请求头
		fwrite($dst, $req);

		if (is_resource($src)) {
			while (!feof($src)) {
				$data = fread($src, 4096);
				if ($data !== false) {
					$ok = $this->_write($dst, $data);
					if ($ok == false) { return false; }
				}
			}
		} else {
			$ok = $this->_write($dst, $src);
			if ($ok == false) { return false; }
		}

		//http结尾空行
		$ok = $this->_write($dst, "");
		if ($ok == false) { return false; }

		list($code, $head, $msg) = $this->_getResponse($dst, "1.1");
		if ($code != '200') {
			trigger_error("PUT {$file} Fail {$msg}");
			return false;
		}

		return true;
	}

	//http分段写
	private function _write($fp, $data)
	{
		//http分段格式
		$len = strlen($data);
		$data = dechex($len) . "\r\n" . $data . "\r\n";

		$n = fwrite($fp, $data);
		if ($n !== strlen($data)) {
			fclose($fp);
			return false;
		}
		return true;
	}

	//得到响应信息
	private function _getResponse($dst, $http="1.0", $return=false)
	{
		$status = stream_get_line($dst, 1024, "\r\n");
		list($code) = sscanf($status, "HTTP/{$http} %d");

		$head = array();
		while ($tmp = stream_get_line($dst, 1024, "\r\n")) {
			list($key, $val) = explode(":", $tmp, 2);
			$head[trim($key)] = trim($val);
		}

		if ($return) {
			return array($code, $head, $dst);
		}

		$msg = stream_get_contents($dst);

		fclose($dst);

		return array($code, $head, $msg);
	}

	private function _parseFile($file)
	{
		return $this->_fix_path($file);
	}

	//修正路径
	private function _fix_path($path)
	{
		$tmp = explode('/', $path);

		$new = array();

		foreach ($tmp as $part) {
			switch ($part) {
			case '':
			case '.':
				break;
			case '..':
				array_pop($new);
				break;
			default:
				$new[] = $part;
				break;
			}
		}

		return implode('/', $new);
	}

}
