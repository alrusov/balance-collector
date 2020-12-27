{{define "dashboard"}}
{{template "header" .}}

<input type="hidden" id="json" name="json" value="" />
<input type="hidden" id="id"   name="id"   value="" />

<table class="grd nobr">
  <tr>
    <th class="right" colspan="{{.Data.ColsCount}}"><button class="micro" style="margin-bottom: 5px;" name="op" value="" onclick="document.getElementById('json').value='y';">JSON</button></th>
  </tr>
  <tr>
    <th>Наименование</th>
    <th>Идентификатор</th>
    <th colspan={{.Data.Fcount}}>Значения</th>
    <th colspan={{.Data.Fcount}}>Изменения</th>
    <th colspan={{.Data.Scount}}>Комментарии</th>
    <th>Обновлено</th>
  </tr>
  {{$replyNeeded := false}}
  {{range $li, $lv := .Data.List}}
    <tr>
      <th class="left normal"><button class="row" name="op" value="history" onclick="document.getElementById('id').value='{{.ID}}'" title="Посмотреть иcторию">{{.Name}}</button></th>
      <th class="left normal"><button class="row" name="op" value="update"  onclick="if(confirm('Получить новые данные для {{.Name}}?')) { document.getElementById('id').value='{{.ID}}'; return true; } else { return false; }" title="Получить новые данные">{{if .Login}}{{.Login}}{{else}}{{.Name}}{{end}}</button></th>

      {{range $i, $v := .Info.FVals}}
        <td class="right{{if index $lv.Ferror $i}} attention{{end}}" title="{{index $lv.FLegend $i}}{{if index $lv.Ferror $i}} {{index $lv.Ferror $i}}{{end}}">{{printf "%.2f" $v}}</td>
      {{end}}
      {{if .Ftail}}
        <td colspan="{{.Ftail}}">&nbsp;</td>
      {{end}}

      {{range $i, $v := .LastChange}}
        <td class="right" title="{{index $lv.FLegend $i}}">{{printf "%.2f" .}}</td>
      {{end}}
      {{if .Ftail}}
        <td colspan="{{.Ftail}}">&nbsp;</td>
      {{end}}

      {{range $i, $v := .Info.SVals}}
        <td title="{{index $lv.SLegend $i}}">{{.}}</td>
      {{end}}
      {{if .Stail}}
        <td colspan="{{.Stail}}">&nbsp;</td>
      {{end}}

      <td class="center{{if .Error}} attention{{end}}"{{if .Error}}{{$replyNeeded = true}} title="{{.Error}}"{{end}}>{{.TS}}</td>
    </tr>
  {{else}}
    <tr>
      <td colspan="6"><strong>Нет данных</strong></td>
    </tr>
  {{end}}
</table>

<table class="buttons-panel">
  <tr>
    <td><button class="command" name="op" value="">Обновить</button></td>
    <td><button class="command" name="op" value="update-all" onclick="return confirm('Получить новые данные для всех?')">Получить новые данные</button></td>
    <td><button class="command{{if not $replyNeeded}} hidden{{end}} " name="op" value="repeat" onclick="return confirm('Повторить неудачные?')">Повторить неудачные</button></td>
  </tr>
</table>


{{template "footer" .}}
{{end}}
