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
	.bar { display: inline-block; height: 30px; min-width: 1px }
	.bar-value { font-size: 70%; line-height: 30px; margin-left: 3px }
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
				<table class="table table-bordered">
					<tr>
						<th>Board</th>
						<th>Declare compatibility</th>
						<th>Declare compatibility but fail to compile</th>
						<th>Don't declare compatibility but compile successfully</th>
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
							<div class="pass bar" style="width: {{ percent $.NumLibsCompatibilityAsterisk }}"
								data-bs-toggle="tooltip" data-bs-placement="top" title="{{ percent $.NumLibsCompatibilityAsterisk }} of the tested libs declare compatibility with any board (architectures=*)">
								&nbsp;
							</div>
							<div class="pass bar" style="width: {{ percent .ExplicitClaim }}"
								data-bs-toggle="tooltip" data-bs-placement="top" title="{{ percent .ExplicitClaim }} of the tested libs declare compatibility with {{ .Architecture }} explicitly">
								&nbsp;
							</div>
							<span class="bar-value">{{ percent .Claim }}</span>
						</td>
						<td width="25%">
							<div class="fail bar" style="width: {{ percent .FailClaimAsterisk }}"
								data-bs-toggle="tooltip" data-bs-placement="top" title="{{ percent .FailClaimAsterisk }} of the tested libs declare generic compatibility (architectures=*) but fail compilation (simple inclusion, not considering examples)">
								&nbsp;
							</div>
							<div class="fail bar" style="width: {{ percent .FailExplicitClaim }}"
								data-bs-toggle="tooltip" data-bs-placement="top" title="{{ percent .FailExplicitClaim }} of the tested libs declare explicit compatibility with {{ .Architecture }} but fail compilation (simple inclusion, not considering examples)">
								&nbsp;
							</div>
							<span class="bar-value">{{ percent .FailClaim }}</span>
						</td>
						<td width="25%">
							<div class="warning bar" style="width: {{ percent .PassNoClaim }}"
								data-bs-toggle="tooltip" data-bs-placement="top" title="{{ percent .PassNoClaim }} don't declare compatibility but can be still compiled for {{ .Name }}">
								&nbsp;
							</div>
							<span class="bar-value">{{ percent .PassNoClaim }}</span>
						</td>
						{{ if $.HasUntested }}
						<td class="text-end">
							{{ .Untested }} ({{ percent .Untested }})
						</td>
						{{ end }}
					</tr>
					{{ end }}
				</table>

				<div class="row">
					<div class="w-25">
						<div class="card border-success mb-3" style="max-width: 18rem;">
							<div class="card-body text-success">
								<h3 class="card-title">{{ percent .NumLibsCompatibilityAsterisk }}</h3>
								<p class="card-text">Declare compatibility with <b>any</b> board (architectures=*).</p>
							</div>
						</div>
					</div>
					<div class="w-25">
						<div class="card border-danger mb-3" style="max-width: 18rem;">
							<div class="card-body text-danger">
								<h3 class="card-title">{{ percent .NumLibsAsteriskFail }}</h3>
								<p class="card-text">Declare compatibility with <b>any</b> board (architectures=*) but had compilation failures.</p>
							</div>
						</div>
					</div>
				</div>

				<div class="row">
					<div class="w-25">
						<div class="card border-success mb-3" style="max-width: 18rem;">
							<div class="card-body text-success">
								<h3 class="card-title">{{ percent .NumLibsClaimAllBoards }}</h3>
								<p class="card-text">Declare compatibility with <b>all</b> the tested boards.</p>
							</div>
						</div>
					</div>
					<div class="w-25">
						<div class="card border-danger mb-3" style="max-width: 18rem;">
							<div class="card-body text-danger">
								<h3 class="card-title">{{ percent .NumLibsClaimNoBoards }}</h3>
								<p class="card-text">Don't declare compatibility with <b>any</b> of the tested boards.</p>
							</div>
						</div>
					</div>
				</div>

				<div class="row">
					<div class="w-25">
						<div class="card border-success mb-3" style="max-width: 18rem;">
							<div class="card-body text-success">
								<h3 class="card-title">{{ percent .NumLibsPassAllBoards }}</h3>
								<p class="card-text">Compile for <b>all</b> the tested boards.</p>
							</div>
						</div>
					</div>
					<div class="w-25">
						<div class="card border-danger mb-3" style="max-width: 18rem;">
							<div class="card-body text-danger">
								<h3 class="card-title">{{ percent .NumLibsPassNoBoards }}</h3>
								<p class="card-text">Do not compile for <b>any</b> of the tested boards.</p>
							</div>
						</div>
					</div>
					<div class="w-25">
						<div class="card border-warning mb-3" style="max-width: 18rem;">
							<div class="card-body text-warning">
								<h3 class="card-title">{{ percent .NumLibsFailClaim }}</h3>
								<p class="card-text">Do not compile on boards they declare to be compatible with.</p>
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
								{{ percent .Count }}
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
				<pre class="pre-scrollable">{{ printf "%.10000s" $t.Log }}</pre>

					{{ range $e := $t.Examples }}
					<h4>examples/{{ $e.Name }}</h4>
					<p>
						Result: <b>{{ $e.Result }}</b>
					</p>
					<pre class="pre-scrollable">{{ printf "%.10000s" $e.Log }}</pre>
					{{ end }}
				{{ end }}
			</div>
		</div>
	</div>
  </body>
</html>
`
