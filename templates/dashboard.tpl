{{define "dashboard"}}
{{template "header" .}}

<input type="hidden" id="id"  name="id"  value="" />
<input type="hidden" id="raw" name="raw" value="" />

<table id="dashboard" class="grd nobr">
  <tr>
    <th class="right nobr" colspan="{{$.Data.ColsCount}}">
      <button class="micro" style="margin: 5p ;" name="op" value="" onclick="document.getElementById('raw').value='json';">JSON</button>
    </th>
  </tr>

  <tr>
    <th>Наименование</th>
    <th>Идентификатор</th>
    <th colspan={{$.Data.Fcount}}>Значения</th>
    <th colspan={{$.Data.Fcount}}>Изменения</th>
    <th colspan={{$.Data.Scount}}>Комментарии</th>
    <th>Обновлено</th>
  </tr>
  {{$replyNeeded := false}}
  {{range $li, $lv := $.Data.List}}
    <tr>
      <th class="left normal"><button class="row" name="op" value="history" onclick="document.getElementById('id').value='{{$lv.ID}}'" title="Посмотреть иcторию">{{$lv.Name}}</button></th>
      <th class="left normal"><button class="row" name="op" value="update"  onclick="if(confirm('Получить новые данные для {{$lv.Name}}?')) { document.getElementById('id').value='{{$lv.ID}}'; return true; } else { return false; }" title="Получить новые данные">{{if $lv.Login}}{{$lv.Login}}{{else}}{{$lv.Name}}{{end}}</button></th>

      {{range $i, $v := $lv.Info.FVals}}
        <td class="right{{if index $lv.Ferror $i}} attention{{end}}" title="{{index $lv.FLegend $i}}{{if index $lv.Ferror $i}} {{index $lv.Ferror $i}}{{end}}">{{printf "%.2f" $v}}</td>
      {{end}}
      {{if $lv.Ftail}}
        <td colspan="{{$lv.Ftail}}">&nbsp;</td>
      {{end}}

      {{range $i, $v := $lv.LastChange}}
        <td class="right" title="{{index $lv.FLegend $i}}">{{printf "%.2f" $v}}</td>
      {{end}}
      {{if $lv.Ftail}}
        <td colspan="{{$lv.Ftail}}">&nbsp;</td>
      {{end}}

      {{range $i, $v := $lv.Info.SVals}}
        <td title="{{index $lv.SLegend $i}}">{{$v}}</td>
      {{end}}
      {{if $lv.Stail}}
        <td colspan="{{$lv.Stail}}">&nbsp;</td>
      {{end}}

      <td class="center{{if $lv.Error}} attention{{end}}"{{if $lv.Error}}{{$replyNeeded = true}} title="{{$lv.Error}}"{{end}}>{{$lv.TS}}</td>
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
