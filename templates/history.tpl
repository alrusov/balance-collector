{{define "history"}}
{{template "header" .}}

<table class="comment nobr" style="width: 1px; margin-bottom: 15px;">
  <tr><th>Наименование:</th><td>{{$.Data.Data.Name}}</td></tr>
  <tr><th>Идентификатор:</th><td>{{$.Data.Data.Login}}</td></tr>
</table>

{{template "back" .}}

<input type="hidden" id="json" name="json" value="" />
<input type="hidden" id="id"   name="id"   value="{{$.Data.Data.ID}}" />

<table class="grd nobr"> 
  <tr>
    <th class="right" colspan="{{$.Data.ColsCount}}"><button class="micro" style="margin-bottom: 5px;" name="op" value="history" onclick="document.getElementById('json').value='y';">JSON</button></th>
  </tr>

  <tr>
    <th>Время</th>
    {{range $.Data.Data.FLegend}}
      <th>{{.}}</th>
    {{end}}
    {{range $.Data.Data.SLegend}}
      <th>{{.}}</th>
    {{end}}
  </tr>

  {{range $li, $lv := $.Data.Data.List}}
    <tr>
      <td class="center">{{$lv.TS}}</td>
      {{range $i, $v := $lv.Info.FVals}}
        <td class="right" title="{{index $.Data.Data.FLegend $i}}">{{printf "%.2f" $v}}</td>
      {{end}}
      {{if index $.Data.Data.Ftail $li}}
        <td colspan="{{index $.Data.Data.Ftail $li}}">&nbsp;</td>
      {{end}}

      {{range $i, $v := $lv.Info.SVals}}
        <td class="left" title="{{index $.Data.Data.SLegend $i}}">{{$v}}</td>
      {{end}}
      {{if index $.Data.Data.Stail $li}}
        <td colspan="{{index $.Data.Data.Stail $li}}">&nbsp;</td>
      {{end}}
    </tr>
  {{else}}
    <tr>
      <td colspan="6"><strong>Нет данных</strong></td>
    </tr>
  {{end}}
</table>

{{template "back" .}}

{{template "footer" .}}
{{end}}
