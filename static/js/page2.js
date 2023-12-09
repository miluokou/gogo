// 在第二页中显示城市基础信息
var cityInfo = [
    { label: '城市', value: '城市名称' },
    { label: '人口', value: '城市人口数量' },
    { label: '城市面积', value: '城市面积大小' },
    { label: '人均GDP', value: '城市人均GDP数值' },
    { label: '失业率', value: '城市失业率百分比' },
    { label: '教育水平', value: '城市受教育程度指标' },
    { label: '消费能力', value: '城市消费能力数据' },
    { label: '基础设施', value: '城市基础设施建设指数' },
    { label: '绿化覆盖率', value: '城市绿化覆盖比例' },
];

var cityTable = document.getElementById('cityTable'); // 获取表格元素

for (var i = 0; i < cityInfo.length; i++) {
    var row = cityTable.insertRow();
    var labelCell = row.insertCell(0);
    var valueCell = row.insertCell(1);
    labelCell.textContent = cityInfo[i].label;
    valueCell.textContent = cityInfo[i].value;

    // 添加单元格边框样式
    labelCell.style.border = '1px solid #333333';
    valueCell.style.border = '1px solid #333333';
}

// 在第二页中展示地图
var map = new AMap.Map('mapContainer', {
    zoom: 12,
});

// 获取北京市范围坐标
AMap.plugin('AMap.DistrictSearch', function () {
    var district = new AMap.DistrictSearch({
        extensions: 'all',
        subdistrict: 0 // 不需要返回下级行政区
    });
    district.search('北京市', function (status, result) {
        if (status === 'complete') {
            var cityBounds = result.districtList[0].boundaries; // 获取北京市范围坐标

            // 创建城市范围围栏
            var cityPolygon = new AMap.Polygon({
                path: cityBounds,
                strokeColor: "#FF33FF",
                strokeOpacity: 0.2,
                strokeWeight: 3,
                fillColor: "#1791fc",
                fillOpacity: 0.4,
            });
            cityPolygon.setMap(map);

            // 调整地图视野以显示围栏数据
            map.setBounds(cityPolygon.getBounds());
        }
    });
});

