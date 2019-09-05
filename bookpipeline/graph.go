package bookpipeline

import (
	"fmt"
	"io"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

const maxticks = 20
const goodCutoff = 70
const mediumCutoff = 65
const badCutoff = 60

type Conf struct {
	Path, Code string
	Conf       float64
}

type GraphConf struct {
	Pgnum, Conf float64
}

func createLine(xvalues []float64, y float64, c drawing.Color) chart.ContinuousSeries {
	var yvalues []float64
	for _ = range xvalues {
		yvalues = append(yvalues, y)
	}
	return chart.ContinuousSeries{
		XValues: xvalues,
		YValues: yvalues,
		Style: chart.Style{
			Show:            true,
			StrokeColor:     c,
			StrokeDashArray: []float64{10.0, 5.0},
		},
	}
}

func Graph(confs map[string]*Conf, bookname string, w io.Writer) error {
	// Organise confs to sort them by page
	var graphconf []GraphConf
	for _, conf := range confs {
		name := filepath.Base(conf.Path)
		var numend int
		numend = strings.Index(name, "_")
		if numend == -1 {
			numend = strings.Index(name, ".")
		}
		pgnum, err := strconv.ParseFloat(name[0:numend], 64)
		if err != nil {
			continue
		}
		var c GraphConf
		c.Pgnum = pgnum
		c.Conf = conf.Conf
		graphconf = append(graphconf, c)
	}
	sort.Slice(graphconf, func(i, j int) bool { return graphconf[i].Pgnum < graphconf[j].Pgnum })

	// Create main xvalues and yvalues, annotations and ticks
	var xvalues, yvalues []float64
	var annotations []chart.Value2
	var ticks []chart.Tick
	i := 0
	tickevery := len(graphconf) / maxticks
	for _, c := range graphconf {
		i = i + 1
		xvalues = append(xvalues, c.Pgnum)
		yvalues = append(yvalues, c.Conf)
		if c.Conf < goodCutoff {
			annotations = append(annotations, chart.Value2{Label: fmt.Sprintf("%.0f", c.Pgnum), XValue: c.Pgnum, YValue: c.Conf})
		}
		if tickevery%i == 0 {
			ticks = append(ticks, chart.Tick{c.Pgnum, fmt.Sprintf("%.0f", c.Pgnum)})
		}
	}
	mainSeries := chart.ContinuousSeries{
		XValues: xvalues,
		YValues: yvalues,
	}

	// Create lines
	goodCutoffSeries := createLine(xvalues, goodCutoff, chart.ColorAlternateGreen)
	mediumCutoffSeries := createLine(xvalues, mediumCutoff, chart.ColorOrange)
	badCutoffSeries := createLine(xvalues, badCutoff, chart.ColorRed)

	// Create lines marking top and bottom 10% confidence
	sort.Slice(graphconf, func(i, j int) bool { return graphconf[i].Conf < graphconf[j].Conf })
	lowconf := graphconf[int(len(graphconf)/10)].Conf
	highconf := graphconf[int((len(graphconf)/10)*9)].Conf
	yvalues = []float64{}
	for _ = range graphconf {
		yvalues = append(yvalues, lowconf)
	}
	minSeries := &chart.ContinuousSeries{
		Style: chart.Style{
			Show:            true,
			StrokeColor:     chart.ColorAlternateGray,
			StrokeDashArray: []float64{5.0, 5.0},
		},
		XValues: xvalues,
		YValues: yvalues,
	}
	yvalues = []float64{}
	for _ = range graphconf {
		yvalues = append(yvalues, highconf)
	}
	maxSeries := &chart.ContinuousSeries{
		Style: chart.Style{
			Show:            true,
			StrokeColor:     chart.ColorAlternateGray,
			StrokeDashArray: []float64{5.0, 5.0},
		},
		XValues: xvalues,
		YValues: yvalues,
	}

	graph := chart.Chart{
		Title:      bookname,
		TitleStyle: chart.StyleShow(),
		Width:      1920,
		Height:     1080,
		XAxis: chart.XAxis{
			Name:      "Page number",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
			Range: &chart.ContinuousRange{
				Min: 0.0,
			},
			Ticks: ticks,
		},
		YAxis: chart.YAxis{
			Name:      "Confidence",
			NameStyle: chart.StyleShow(),
			Style:     chart.StyleShow(),
			Range: &chart.ContinuousRange{
				Min: 0.0,
				Max: 100.0,
			},
		},
		Series: []chart.Series{
			mainSeries,
			minSeries,
			maxSeries,
			goodCutoffSeries,
			mediumCutoffSeries,
			badCutoffSeries,
			chart.AnnotationSeries{
				Annotations: annotations,
			},
		},
	}
	return graph.Render(chart.PNG, w)
}
