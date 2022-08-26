package application

import (
	"github.com/coroot/coroot-focus/model"
	"github.com/coroot/coroot-focus/timeseries"
	"github.com/coroot/coroot-focus/views/widgets"
)

func cpu(app *model.Application) *widgets.Dashboard {
	dash := &widgets.Dashboard{Name: "CPU"}
	relevantNodes := map[string]*model.Node{}

	for _, i := range app.Instances {
		for _, c := range i.Containers {
			dash.GetOrCreateChartInGroup("CPU usage of container <selector>, cores", c.Name).
				AddSeries(i.Name, c.CpuUsage).
				SetThreshold("limit", c.CpuLimit, timeseries.Max)
			dash.GetOrCreateChartInGroup("CPU delay of container <selector>, seconds/second", c.Name).AddSeries(i.Name, c.CpuDelay)
			dash.GetOrCreateChartInGroup("Throttled time of container <selector>, seconds/second", c.Name).AddSeries(i.Name, c.ThrottledTime)
		}
		if node := i.Node; i.Node != nil {
			nodeName := node.Name.Value()
			if relevantNodes[nodeName] == nil {
				relevantNodes[nodeName] = i.Node
				dash.GetOrCreateChartInGroup("Node CPU usage <selector>, %", "overview").
					AddSeries(nodeName, i.Node.CpuUsagePercent).
					Feature()

				byMode := dash.GetOrCreateChartInGroup("Node CPU usage <selector>, %", nodeName).Sorted().Stacked()
				for _, s := range cpuByMode(node.CpuUsageByMode) {
					byMode.Series = append(byMode.Series, s)
				}

				usageByApp := map[string]timeseries.TimeSeries{}
				for _, instance := range node.Instances {
					appUsage := usageByApp[instance.OwnerId.Name]
					if appUsage == nil {
						appUsage = timeseries.Aggregate(timeseries.NanSum)
						usageByApp[instance.OwnerId.Name] = appUsage
					}
					for _, c := range instance.Containers {
						appUsage.(*timeseries.AggregatedTimeseries).AddInput(c.CpuUsage)
					}
				}
				dash.GetOrCreateChartInGroup("CPU consumers on <selector>, cores", nodeName).
					Stacked().
					Sorted().
					SetThreshold("total", node.CpuCapacity, timeseries.Any).
					AddMany(timeseries.Top(usageByApp, timeseries.NanSum, 5))
			}
		}
	}
	return dash
}

func cpuByMode(modes map[string]timeseries.TimeSeries) []*widgets.Series {
	var res []*widgets.Series
	for _, mode := range []string{"user", "nice", "system", "wait", "iowait", "steal", "irq", "softirq"} {
		v, ok := modes[mode]
		if !ok {
			continue
		}
		var color string
		switch mode {
		case "user":
			color = "blue"
		case "system":
			color = "red"
		case "wait", "iowait":
			color = "orange"
		case "steal":
			color = "black"
		case "irq":
			color = "grey"
		case "softirq":
			color = "yellow"
		case "nice":
			color = "lightGreen"
		}
		res = append(res, &widgets.Series{Name: mode, Color: color, Data: v})
	}
	return res
}