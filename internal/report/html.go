package report

var htmlTmpl = `
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.2/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-EVSTQN3/azprG1Anm3QDgpJLIm9Nao0Yz1ztcQTwFspd3yD65VohhpuuCOmLASjC" crossorigin="anonymous">
    <title>arduino-testlib report</title>
	<style>
	.pass { background-color: #00FF00 !important; }
	.fail { background-color: #FF0000 !important; }
	</style>
  </head>
  <body>
	<div class="container">
		<div class="row">
			<div class="col">
				<h1>Arduino libraries compatibility matrix</h1>
				<p>
					<small>This report was generated on {{ .Timestamp }} using 
					<a href="https://github.com/alranel/arduino-testlib">arduino-testlib</a>.</small>
				</p>

				<h2>Overview</h2>
				<p><b>{{ .NumLibs }}</b> libraries were tested on <b>{{ .NumBoards }} boards.</b></p>

				<h2>Boards</h2>
				<p>The following table represents how many libraries are compatible with each tested board, 
					and how many claim compatibility with its architecture in their library.properties metadata file.</p>
				<table class="table table-bordered">
					<tr>
						<th>Board</th>
						<th class="text-end">Compatible</th>
						<th class="text-end">Incompatible</th>
						<th class="text-end">Untested</th>
					</tr>
					{{ range .Boards }}
					<tr>
						<td>
							<b>{{ .Name }}</b><br />
							<small>{{ .Versions }}</small>
						</td>
						<td class="text-end">
							{{ .TotPass }} ({{ .TotPassPercent }})<br />
							<small>claiming compatibility: {{ .PassClaim }} ({{ .PassClaimPercent }})</small>
						</td>
						<td class="text-end">
							{{ .TotFail }} ({{ .TotFailPercent }})<br />
							<small>claiming compatibility: {{ .FailClaim }} ({{ .FailClaimPercent }})</small>
						</td>
						<td class="text-end">
							{{ .Untested }} ({{ .UntestedPercent }})
						</td>
					</tr>
					{{ end }}
				</table>
				<p>
					<b>{{ .NumLibsAllBoards }}</b> libraries ({{ .NumLibsAllBoardsPercent }}) are compatible with all the tested boards.<br />
					<b>{{ .NumLibsNoBoards }}</b> libraries ({{ .NumLibsNoBoardsPercent }}) are not compatible with any of them.<br />
				</p>

				<h2>Libraries</h2>
				<p>The following table shows the compatibility status for each library.</p>
				<p>The ⚠️ icon indicates those cases when compilation status does not match the compatibility
					advertised in library.properties.</p>
				<table class="table table-bordered table-sm">
					<tr>
						<th>Library</th>
						<th>Version</th>
						{{ range .Boards }}
						<th>{{ .Name }}</th>
						{{ end }}
					</tr>
					{{ range $lib := .Libraries }}
					<tr>
						<td>
							<a href="{{ $lib.ReportFile }}"><b>{{ $lib.Name }}</b></a>
						</td>
						<td>{{ .Version }}</td>
						{{ range $board := $.Boards }}
						{{ if eq (index $lib.BoardCompatibility $board.Name) "PASS_CLAIM" }}<td class="pass">PASS</td>{{ end }}
						{{ if eq (index $lib.BoardCompatibility $board.Name) "PASS_NOCLAIM" }}<td class="pass">PASS ⚠️</td>{{ end }}
						{{ if eq (index $lib.BoardCompatibility $board.Name) "FAIL_CLAIM" }}<td class="fail">FAIL ⚠️</td>{{ end }}
						{{ if eq (index $lib.BoardCompatibility $board.Name) "FAIL_NOCLAIM" }}<td class="fail">FAIL</td>{{ end }}
						{{ if eq (index $lib.BoardCompatibility $board.Name) "" }}<td></td>{{ end }}
						{{ end }}
					</tr>
					{{ end }}
				</table>

				<h2>Misc stats</h2>
				<h3>Examples</h3>
				<p>The following table shows how many examples are provided per library.</p>
				<table class="table table-bordered">
					<tr>
						<th>Number of examples</th>
						<th>Number of libraries</th>
					</tr>
					{{ range .Examples }}
					<tr>
						<td>
							<b>{{ .Num }}</b>
						</td>
						<td>
							{{ .Count }} ({{ .CountPercent }})
						</td>
					</tr>
					{{ end }}
				</table>
			</div>
		</div>
	</div>
  </body>
</html>
`

var htmlTmplLibrary = `
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.2/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-EVSTQN3/azprG1Anm3QDgpJLIm9Nao0Yz1ztcQTwFspd3yD65VohhpuuCOmLASjC" crossorigin="anonymous">
    <title>arduino-testlib report</title>
	<style>
	.pass { background-color: #00FF00 !important; }
	.fail { background-color: #FF0000 !important; }
	</style>
  </head>
  <body>
	<div class="container">
		<div class="row">
			<div class="col">
				<h1>{{ .Lib.Name }} - compatibility matrix</h1>
				<p>
					<small>This report was generated on {{ .Timestamp }} using 
					<a href="https://github.com/alranel/arduino-testlib">arduino-testlib</a>.</small>
				</p>

				<h2>{{ .Lib.Name }}</h2>
				<p>
					Version: <b>{{ .Lib.Version }}</b>
					<br />
					<a href="{{ .Lib.URL }}">More details</a>
				</p>

				<h2>Compatibility matrix</h2>
				<table class="table table-bordered">
					<tr>
						<th>Board</th>
						<th>Claims compatibility</th>
						<th>Inclusion</th>
						{{ range $e := $.Lib.Examples }}
						<th><pre>{{ $e }}</pre></th>
						{{ end }}
					</tr>
					{{ range $board := .Boards }}
					<tr>
						<td>
							<a href="#{{ $board.Name }}"><b>{{ $board.Name }}</b></a>
							<br />
							<small>{{ $board.Versions }}</small>
						</td>

						{{ if eq (index $.Lib.BoardCompatibility $board.Name) "PASS_CLAIM" }}
							<td>Yes</td>
							<td class="pass">PASS</td>
						{{ end }}
						{{ if eq (index $.Lib.BoardCompatibility $board.Name) "PASS_NOCLAIM" }}
							<td>No ⚠️</td>
							<td class="pass">PASS</td>
						{{ end }}
						{{ if eq (index $.Lib.BoardCompatibility $board.Name) "FAIL_CLAIM" }}
							<td>Yes ⚠️</td>
							<td class="fail">FAIL</td>
						{{ end }}
						{{ if eq (index $.Lib.BoardCompatibility $board.Name) "FAIL_NOCLAIM" }}
							<td>No</td>
							<td class="fail">FAIL</td>
						{{ end }}

						{{ range $e := $.Lib.Examples }}
							{{ range $t := $.Lib.BoardTestResults }}
								{{ if eq $t.FQBN $board.Name }}
									{{ range $te := $t.Examples }}
										{{ if eq $te.Name $e }}
											{{ if eq $te.Result "PASS" }}<td class="pass">PASS</td>{{ end }}
											{{ if eq $te.Result "FAIL" }}<td class="fail">FAIL</td>{{ end }}
										{{ end }}
									{{ end }}
								{{ end }}
							{{ end }}
						{{ end }}
					</tr>
					{{ end }}
				</table>

				<h2>Compilation logs</h2>
				{{ range $t := $.Lib.BoardTestResults }}
				<h3 id="{{ $t.FQBN }}">{{ $t.FQBN }} @ {{ $t.CoreVersion }}</h3>
				<h4>Inclusion</h4>
				<p>
					Result: <b>{{ $t.Result }}</b>
					{{ if $t.NoMainHeader }}
					<br />This library has no main header file so an empty one was created.
					{{ end }}
				</p>
				<pre class="pre-scrollable">{{ $t.Log }}</pre>

					{{ range $e := $t.Examples }}
					<h4>examples/{{ $e.Name }}</h4>
					<p>
						Result: <b>{{ $e.Result }}</b>
					</p>
					<pre class="pre-scrollable">{{ $e.Log }}</pre>
					{{ end }}
				{{ end }}
			</div>
		</div>
	</div>
  </body>
</html>
`
