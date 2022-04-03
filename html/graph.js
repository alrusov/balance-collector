//----------------------------------------------------------------------------------------------------------------------------//

var options = {
    format: '#,###.00',
    colors: ['#000080', '#008000', '#800000', '#808080', '#000008', '#000800', '#080000', '#080808'],
}

//----------------------------------------------------------------------------------------------------------------------------//

am5.ready(
    function () {
        // Корень
        var root = am5.Root.new("chartdiv");

        //----------------------------------------------------------------------------------------------------------------------------//

        root.setThemes([
            // Анимация. Лучше отключим
            //  am5themes_Animated.new(root)
        ]);

        // Добавляем чарт в контейнер
        var chart = root.container.children.push(
            am5xy.XYChart.new(
                root,
                {
                    focusable: true,
                    panX: true,
                    panY: false,
                    wheelX: "panX",
                    wheelY: "zoomX",
                    layout: root.verticalLayout
                }
            )
        );

        // График будет линейным
        var easing = am5.ease.linear;

        // Ось X (время), она общая
        var xAxis = chart.xAxes.push(
            am5xy.DateAxis.new(
                root,
                {
                    maxDeviation: 0.1,
                    groupData: false, // если сделать true, то слетает автомасштабирование по Y, да вообще группировка нам не нужна
                    baseInterval: {
                        timeUnit: "day",
                        count: 1
                    },
                    renderer: am5xy.AxisRendererX.new(
                        root,
                        {
                            minGridDistance: 100
                        }
                    ),
                }
            )
        );

        //----------------------------------------------------------------------------------------------------------------------------//

        // Добавление серии с индивидуальной вертикальной осью
        //   name - имя
        //   isRightAxis - ось справа?
        function createAxisAndSeries(name, isRightAxis) {
            // Ось
            var yRenderer = am5xy.AxisRendererY.new(
                root,
                {
                    opposite: isRightAxis
                }
            );
            var yAxis = chart.yAxes.push(
                am5xy.ValueAxis.new(
                    root,
                    {
                        maxDeviation: 1,
                        renderer: yRenderer,
                    }
                )
            );

            // Синхронизируем обновление всех осей с первой
            if (chart.yAxes.indexOf(yAxis) > 0) {
                yAxis.set("syncWithAxis", chart.yAxes.getIndex(0));
            }

            var idx = chart.series.length;

            // Добавляем серию
            var serie = chart.series.push(
                am5xy.LineSeries.new(
                    root,
                    {
                        name: name,
                        xAxis: xAxis,
                        yAxis: yAxis,
                        valueYField: "value",
                        valueXField: "date",

                        tooltip: am5.Tooltip.new(
                            root,
                            {
                                pointerOrientation: "horizontal",
                                labelText: "[bold]{name}[/]\n{valueX.formatDate('yyyy-MM-dd')}\n{valueY.formatNumber('" + options.format + "')}"
                            }
                        ),

                        legendLabelText: "[{stroke}]{name}: [/]",
                        legendRangeLabelText: "[{stroke}]{name}[/]",
                        legendValueText: "[{stroke}]{valueY.formatNumber('" + options.format + "')}[/]",
                        legendRangeValueText: "[{stroke}]{valueYClose.formatNumber('" + options.format + "')}[/]", // не работает
                    }
                )
            );

            var color = am5.color(options.colors[idx % options.colors.length]);

            serie.setAll(
                {
                    "fill": color, // тут это влияет на цвет заливки tooltip
                    "stroke": color, // цвет линии
                }
            );

            serie.bullets.push(function () {
                return am5.Bullet.new(root, {
                    sprite: am5.Circle.new(root, {
                        radius: 1,
                        fill: serie.get("fill"),
                        stroke: serie.get("stroke"), //root.interfaceColors.get("background"),
                        strokeWidth: 2
                    })
                });
            });

            serie.strokes.template.setAll(
                {
                    strokeWidth: 1 // толщина линии
                }
            );

            yRenderer.grid.template.set("strokeOpacity", 0.05); // прозрачность горизонтального грида

            yRenderer.labels.template.set("fill", color); // цвет цифр на оси

            yRenderer.setAll(
                {
                    stroke: color, // цвет оси
                    strokeOpacity: 0.5, // и её прозрачность
                }
            );

            // Как брать горизонтальные данные
            serie.data.processor = am5.DataProcessor.new(
                root,
                {
                    dateFormat: "yyyy-MM-dd",
                    dateFields: ["date"]
                }
            );

            return serie;
        }

        //----------------------------------------------------------------------------------------------------------------------------//

        // Курсор

        var cursor = chart.set(
            "cursor",
            am5xy.XYCursor.new(
                root,
                {
                    xAxis: xAxis,
                    behavior: "none"
                }
            )
        );

        cursor.lineX.set("visible", true);
        cursor.lineY.set("visible", true);

        //----------------------------------------------------------------------------------------------------------------------------//

        // Скролбар
        var scrollbarX = am5xy.XYChartScrollbar.new(
            root,
            {
                orientation: "horizontal",
                height: 50
            }
        );

        // Цепляем на график снизу
        chart.set("scrollbarX", scrollbarX);
        chart.bottomAxesContainer.children.push(scrollbarX);

        // Делаем отображение на скролбаре контура первого графика (привязка sbserie внизу при отрисовке)
        var sbxAxis = scrollbarX.chart.xAxes.push(
            am5xy.DateAxis.new(
                root,
                {
                    groupData: false, // если сделать true, то слетает автомасштабирование по Y, да вообще группировка нам не нужна
                    baseInterval: { timeUnit: "day", count: 1 },
                    renderer: am5xy.AxisRendererX.new(
                        root,
                        {
                            opposite: false,
                            strokeOpacity: 0 // ось X в скролбаре не рисуем
                        }
                    )
                }
            )
        );

        // ось Y в скролбаре
        var sbyAxis = scrollbarX.chart.yAxes.push(
            am5xy.ValueAxis.new(
                root,
                {
                    renderer: am5xy.AxisRendererY.new(
                        root,
                        {}
                    )
                }
            )
        );

        // Привязываем контурный график к скролбару
        var sbserie = scrollbarX.chart.series.push(
            am5xy.LineSeries.new(
                root,
                {
                    xAxis: sbxAxis,
                    yAxis: sbyAxis,
                    valueYField: "value",
                    valueXField: "date"
                }
            )
        );

        //----------------------------------------------------------------------------------------------------------------------------//

        // Парсим данные
        var data = JSON.parse(jdata);

        // Создаем серии
        var series = new Array();

        data.fLegend.forEach(
            function (name, idx) {
                series.push(createAxisAndSeries(name, idx % 2 != 0)); // ось Y для четных справа, нечетных слева
            }
        );

        // Создаем и добавляем легенду снизу (получается из-за этого места в коде) посередине (centerX и x 50%)
        var legend = chart.children.push(
            am5.Legend.new(
                root,
                {
                    centerX: am5.p50,
                    x: am5.p50,
                    oversizedBehavior: "wrap",
                    maxWidth: 200,
                    textAlign: "center",
                }
            )
        );

        legend.data.setAll(chart.series.values);

        //----------------------------------------------------------------------------------------------------------------------------//

        // Готовим данные

        var vals = Array();
        series.forEach(
            function () {
                vals.push(new Array())
            }
        );

        data.list.forEach(
            function (x) {
                var ts = new Date(x.ts).getTime();

                x.data.fVals.forEach(
                    function (v, idx) {
                        vals[idx].push(
                            {
                                date: ts,
                                value: v
                            }
                        );
                    }
                )
            }
        );

        // Ну и собственно обновление
        vals.forEach(
            function (serie, idx) {
                // отображаем
                series[idx].data.setAll(serie);

                if (idx == 0) {
                    // для первой серии отображаем на скролбар
                    sbserie.data.setAll(serie);
                }
            }
        );

        //----------------------------------------------------------------------------------------------------------------------------//

    }
); // end am5.ready()
