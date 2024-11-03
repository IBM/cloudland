var myChart = echarts.init(document.getElementById('dashboard'));
var labelTop = {
    normal : {
        label : {
            show : true,
            position : 'center',
            formatter : '{b}',
            textStyle: {
                baseline : 'bottom'
            }
        },
        labelLine : {
            show : false
        }
    }
};
var labelFromatter = {
    normal : {
        label : {
            formatter : function (params){
                res = params.value
                if (params.name.indexOf('volume') > -1 || params.name.indexOf('disk') > -1) {
                    res = res + 'G'
                } else if (params.name.indexOf('memory') > -1) {
                    res = res + 'M'
                }
                res = res + '\n' + params.percent + '%'
                return res
            },
            textStyle: {
                baseline : 'top'
            },
            position: 'outside'
        }
    },
}
var labelBottom = {
    normal : {
        color: '#ccc',
        label : {
            show : true,
            position : 'center'
        },
        labelLine : {
            show : false
        }
    }
};
var radius = [40, 55];
option = {
    title: {
        subtext : 'cpu memory disk volume ip',
        x : 'center',
    },
    legend: {
        x : 'center',
        y : 'center',
        data:[
            'cpu_used','memory_used','disk_used','volume_storage_used','public_ip_used','private_ip_used'
        ]
    },
    toolbox: {
        show : true,
        feature : {
            dataView : {show: true, readOnly: false},
            magicType : {
                show: true, 
                type: ['pie', 'funnel'],
                option: {
                    funnel: {
                        width: '20%',
                        height: '30%',
                        itemStyle : {
                            normal : {
                                label : {
                                    formatter : function (params){
                                        return 'other\n' + params.value + '%\n'
                                    },
                                    textStyle: {
                                        baseline : 'middle'
                                    }
                                }
                            },
                        } 
                    }
                }
            },
            restore : {show: true},
            saveAsImage : {show: true}
        }
    },
    series : [
        {
            name: 'cpu',
            type : 'pie',
            center : ['20%', '30%'],
            avoidLabelOverlap: true,
            radius : radius,
            x: '0%', // for funnel
            name: 'cpu',
            itemStyle : labelFromatter,
            data : [
                {name:'other', itemStyle : labelBottom},
                {name:'cpu_used', itemStyle : labelTop}
            ]
        },
        {
            type : 'pie',
            center : ['50%', '30%'],
            radius : radius,
            x:'20%', // for funnel
            itemStyle : labelFromatter,
            data : [
                {name:'other', itemStyle : labelBottom},
                {name:'memory_used', itemStyle : labelTop}
            ]
        },
        {
            type : 'pie',
            center : ['80%', '30%'],
            radius : radius,
            x:'40%', // for funnel
            itemStyle : labelFromatter,
            data : [
                {name:'other', itemStyle : labelBottom},
                {name:'disk_used', itemStyle : labelTop}
            ]
        },
        {
            type : 'pie',
            center : ['20%', '75%'],
            radius : radius,
            y: '55%',   // for funnel
            x: '0%',    // for funnel
            itemStyle : labelFromatter,
            data : [
                {name:'other', itemStyle : labelBottom},
                {name:'volume_storage_used', itemStyle : labelTop}
            ]
        },
        {
            type : 'pie',
            center : ['50%', '75%'],
            radius : radius,
            y: '55%',   // for funnel
            x:'20%',    // for funnel
            itemStyle : labelFromatter,
            data : [
                {name:'other', itemStyle : labelBottom},
                {name:'public_ip_used', itemStyle : labelTop}
            ]
        },
        {
            type : 'pie',
            center : ['80%', '75%'],
            radius : radius,
            y: '55%',   // for funnel
            x:'20%',    // for funnel
            itemStyle : labelFromatter,
            data : [
                {name:'other', itemStyle : labelBottom},
                {name:'private_ip_used', itemStyle : labelTop}
            ]
        },
{
	    name: 'cpu',
            type:'pie',
            radius: [0, 25],
            color: '#fff',
            center : ['20%', '30%'],
            label: {
                normal: {
            formatter : function (params){
//		console.log(params)
                return params.name + '\n' + params.value
            },
            textStyle: {
                baseline : 'top',
                color: '#000'
            },
		    show: true,
                    position: 'center'
                }
            },
            data:[
                {name:'cpu'},
            ]
        },
{
	    name: 'memory',
            type:'pie',
            radius: [0, 25],
            color: '#fff',
            center : ['50%', '30%'],
            label: {
                normal: {
            formatter : function (params){
//		console.log(params)
                return params.name + '\n' + params.value + 'M'
            },
            textStyle: {
                baseline : 'top',
                color: '#000'
            },
		    show: true,
                    position: 'center'
                }
            },
            data:[
                {name:'memory'},
            ]
        },
{
	    name: 'disk',
            type:'pie',
            radius: [0, 25],
            color: '#fff',
            center : ['80%', '30%'],
            label: {
                normal: {
            formatter : function (params){
//		console.log(params)
                return params.name + '\n' + params.value + 'G'
            },
            textStyle: {
                baseline : 'top',
                color: '#000'
            },
		    show: true,
                    position: 'center'
                }
            },
            data:[
                {name:'disk'},
            ]
        },
{
	    name: 'volume',
            type:'pie',
            radius: [0, 25],
            color: '#fff',
            center : ['20%', '75%'],
            label: {
                normal: {
            formatter : function (params){
//		console.log(params)
                return params.name + '\n' + params.value + 'G'
            },
            textStyle: {
                baseline : 'top',
                color: '#000'
            },
		    show: true,
                    position: 'center'
                }
            },
            data:[
                {name:'volume'},
            ]
        },
{
	    name: 'public_ip',
            type:'pie',
            radius: [0, 25],
            color: '#fff',
            center : ['50%', '75%'],
            label: {
                normal: {
            formatter : function (params){
//		console.log(params)
                return params.name + '\n' + params.value
            },
            textStyle: {
                baseline : 'top',
                color: '#000'
            },
		    show: true,
                    position: 'center'
                }
            },
            data:[
                {name:'public_ip'},
            ]
        },
{
	    name: 'private_ip',
            type:'pie',
            radius: [0, 25],
            color: '#fff',
            center : ['80%', '75%'],
            label: {
                normal: {
            formatter : function (params){
//		console.log(params)
                return params.name + '\n' + params.value
            },
            textStyle: {
                baseline : 'top',
                color: '#000'
            },
		    show: true,
                    position: 'center'
                }
            },
            data:[
                {name:'private_ip'},
            ]
        },
    ]
};

getResourceData();

var int=self.setInterval("getResourceData()",10000);

function getResourceData()
{
$.ajax({
    url: "/dashboard/getdata",
    type: 'GET',
    success: function (data) {
        if (data) {
                option.title.text = data.title
                option.series[0].data[0].value = data.cpu_avail
                option.series[0].data[1].value = data.cpu_used
                option.series[1].data[0].value = data.mem_avail
                option.series[1].data[1].value = data.mem_used
                option.series[2].data[0].value = data.disk_avail
                option.series[2].data[1].value = data.disk_used
                option.series[3].data[0].value = data.volume_avail
                option.series[3].data[1].value = data.volume_used
                option.series[4].data[0].value = data.pubip_avail
                option.series[4].data[1].value = data.pubip_used
                option.series[5].data[0].value = data.prvip_avail
                option.series[5].data[1].value = data.prvip_used
                option.series[6].data[0].value = data.cpu_avail + data.cpu_used
                option.series[7].data[0].value = data.mem_avail + data.mem_used
                option.series[8].data[0].value = data.disk_avail + data.disk_used
                option.series[9].data[0].value = data.volume_avail + data.volume_used
                option.series[10].data[0].value = data.pubip_avail + data.pubip_used
                option.series[11].data[0].value = data.prvip_avail + data.prvip_used
                myChart.setOption(option);
        }
    },
    error: function (jqXHR, textStatus, errorThrown) {
        window.location.href = "/error?ErrorMsg=" + jqXHR.responseText;
    }
});

}




