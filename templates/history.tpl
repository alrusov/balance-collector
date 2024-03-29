{{define "history"}}
{{template "header" .}}

<table class="comment nobr" style="width: 1px; margin-bottom: 15px;">
  <tr><th>Наименование:</th><td>{{$.Data.Data.Name}}</td></tr>
  <tr><th>Идентификатор:</th><td>{{$.Data.Data.Login}}</td></tr>
</table>

{{template "back" .}}

<input type="hidden" id="id"  name="id"  value="{{$.Data.Data.ID}}" />
<input type="hidden" id="tp" name="tp" value="" />

<table id="history" class="grd nobr">
  <tr>
    <th class="right nobr" colspan="{{$.Data.ColsCount}}">
      <button class="micro" style="margin: 5p ;" name="op" value="history" onclick="document.getElementById('tp').value='graph';">График</button>
      <button class="micro" style="margin: 5p ;" name="op" value="history" onclick="document.getElementById('tp').value='json';">JSON</button>
    </th>
  </tr>

  <tr>
    <th>Время</th>
    {{range $.Data.Data.FLegend}}
      <th colspan="2">{{.}}</th>
    {{end}}
    {{range $.Data.Data.SLegend}}
      <th>{{.}}</th>
    {{end}}
  </tr>

  {{range $li, $lv := $.Data.Data.List}}
    <tr>
      <td class="center">{{$lv.TS}}</td>
      {{range $i, $v := $lv.Info.FVals}}
        <td class="right">{{printf "%.2f" $v}}</td>
        {{if (index $lv.Change $i)}}
          <td class="right small">{{printf "%+.2f" (index $lv.Change $i)}}</td>
        {{else}}
          <td>&nbsp;</td>
        {{end}}
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
