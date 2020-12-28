{{define "header"}}
<!DOCTYPE html>
<html lang="ru">
  <head>
    <meta charset="UTF-8" />
    <title>{{$.Name}}</title>
    <link rel="stylesheet" href="{{$.Prefix}}/___.css" />
    <link rel="stylesheet" href="{{$.Prefix}}/css/local.css" />
  </head>

  <body>
    <div id="pls_wait" style="display: none; margin: 15px 0px;">
      <strong class="attention" style="font-size: 20px; border: 1px solid red; padding: 7px;">Выполняется&#8230;</strong>
    </div>

    <h4><img src="/favicon.ico" style="width: 16px; height: 16px; position: relative; top: 2px;" alt="" />&nbsp;<em>{{$.Name}} [{{$.App}} {{$.Version}}{{if $.Tags}}&nbsp;{{$.Tags}}{{end}}]</em></h4>

    <form method="POST" action="{{$.Prefix}}/" onsubmit="
      var f = this;
      setTimeout(
        function() {
          f.style.display='none'; 
          document.getElementById('pls_wait').style.display='block';
          document.body.style.cursor='wait'; 
        },
        300,
      );
      return true;
    ">

    <h1>{{$.Title}}</h1>

    {{if $.Error}}
      <table class="comment nobr" style="width: 1px; margin-bottom: 15px;">
        <tr>
          <th>Ошибка</th>
          <td class="strong attention">{{$.Error}}</td>
        </tr>
      </table>
    {{end}}
{{end}}
