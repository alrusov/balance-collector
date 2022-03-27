{{define "graph"}}
{{template "header" .}}

{{template "back" .}}

<script src="https://code.jquery.com/jquery-3.6.0.min.js"></script>
<script src="https://cdn.amcharts.com/lib/5/index.js"></script>
<script src="https://cdn.amcharts.com/lib/5/xy.js"></script>
<script src="https://cdn.amcharts.com/lib/5/themes/Animated.js"></script>

<script>
var jdata = '{{.Data}}';
</script>

<script src="graph.js"></script>

<div id="chartdiv" style="min-width: 500px; height: 500px; margin: 30px; border: solid 1px gray;"></div>

{{template "back" .}}

{{template "footer" .}}
{{end}}
