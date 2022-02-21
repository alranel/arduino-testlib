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
	.warning { background-color: #FFFF00 !important; }
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
				<p>The following table summarizes how many libraries are compatible with each tested board, 
					and how many libraries have a potentially inaccurate library.properties metadata file 
					in terms of declared compatibility.</p>
				<table class="table table-bordered">
					<tr>
						<th>Board</th>
						<th>Compatibility</th>
						<th>Metadata Mismatch</th>
						{{ if .HasUntested }}
						<th class="text-end">Untested</th>
						{{ end }}
					</tr>
					{{ range .Boards }}
					<tr>
						<td>
							<b>{{ .Name }}</b><br />
							<small>{{ .Versions }}</small>
						</td>
						<td width="25%">
							<div class="pass" style="font-size: 70%; width: {{ .TotPassPercent }}; padding: 3px"
								data-bs-toggle="tooltip" data-bs-placement="top" title="{{ .TotPassPercent }} of the tested libs are compatible with {{ .Name }} (inclusion test only, no examples)">
								{{ .TotPassPercent }}
							</div>
						</td>
						<td width="25%">
							<div class="warning" style="font-size: 70%; width: {{ .TotClaimMismatchPercent }}; padding: 3px"
								data-bs-toggle="tooltip" data-bs-placement="top" title="{{ .FailClaimPercent }} of the tested libs declare compatibility with {{ .Name }} but fail compilation, while {{ .PassNoClaimPercent }} don't declare compatibility but can be compiled">
								{{ .TotClaimMismatchPercent }}
							</div>
						</td>
						{{ if $.HasUntested }}
						<td class="text-end">
							{{ .Untested }} ({{ .UntestedPercent }})
						</td>
						{{ end }}
					</tr>
					{{ end }}
				</table>

				<div class="row row-cols-1 row-cols-md-3 g-4">
					<div class="col">
						<div class="card border-success mb-3" style="max-width: 18rem;">
							<div class="card-body text-success">
								<h3 class="card-title">{{ .NumLibsAllBoardsPercent }}</h3>
								<p class="card-text">Compatible with <b>all</b> boards.</p>
							</div>
						</div>
					</div>
					<div class="col">
						<div class="card border-danger mb-3" style="max-width: 18rem;">
							<div class="card-body text-danger">
								<h3 class="card-title">{{ .NumLibsNoBoardsPercent }}</h3>
								<p class="card-text">Compatible with <b>no</b> boards.</p>
							</div>
						</div>
					</div>
				</div>

				<h2>Libraries</h2>
				<table class="table table-bordered table-sm">
					<tr>
						<th>Library</th>
						<th>Version</th>
						{{ range .Boards }}
						<th><p style="writing-mode: vertical-rl">{{ .Name }}</p></th>
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
						{{ if eq (index $lib.BoardCompatibility $board.Name) "PASS_NOCLAIM" }}<td class="pass" data-bs-toggle="tooltip" data-bs-placement="top" title="This result does not match the architectures declared in library.properties">PASS ⚠️</td>{{ end }}
						{{ if eq (index $lib.BoardCompatibility $board.Name) "FAIL_CLAIM" }}<td class="fail" data-bs-toggle="tooltip" data-bs-placement="top" title="This result does not match the architectures declared in library.properties">FAIL ⚠️</td>{{ end }}
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
						{{ range .Examples }}
							<th>{{ .Num }}</th>
						{{ end }}
					</tr>
					<tr>
						{{ range .Examples }}
							<td style="{{ if eq .Num 0 }}color: red{{ end }}">
								{{ .CountPercent }}
							</td>
						{{ end }}
					</tr>
				</table>
			</div>
		</div>
	</div>
	<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.1.3/dist/js/bootstrap.bundle.min.js" integrity="sha384-ka7Sk0Gln4gmtz2MlQnikT1wXgYsOg+OMhuP+IlRH9sENBO0LRn5q+8nbTov4+1p" crossorigin="anonymous"></script>
	<script>
		var tooltipTriggerList = [].slice.call(document.querySelectorAll('[data-bs-toggle="tooltip"]'));
		var tooltipList = tooltipTriggerList.map(function (tooltipTriggerEl) {
			return new bootstrap.Tooltip(tooltipTriggerEl)
		});
	</script>
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
