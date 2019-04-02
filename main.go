package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"time"

	"go.opencensus.io/exporter/jaeger"
	"go.opencensus.io/trace"
)

func foodHandler(w http.ResponseWriter, r *http.Request) {
	food := r.URL.Path[1:]
	ctx, span := trace.StartSpan(context.Background(), "Handle Food")
	defer span.End()

	for i := 0; i < 5; i++ {
		go eat(ctx, i, food)
	}
	w.Write([]byte("Finished eating\n"))
}

func eat(ctx context.Context, gopher int, food string) {
	// Create child span from the parent context.
	_, childSpan := trace.StartSpan(ctx, "Eat Food")
	defer childSpan.End()

	var msg string
	switch food {
	case "carrot":
		msg = "I like it :)"
	case "fish":
		msg = "I hate it."
		// Gopher hates fish so it's slow to eat.
		time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	default:
		msg = "It's OK."
	}

	// OpenCensus span does not have Log field.
	// Jaeger exporter convert annotations to Logs of OpenTracing span.
	childSpan.Annotate([]trace.Attribute{
		trace.StringAttribute("food", food),
	}, msg)
}

func main() {
	je, err := jaeger.NewExporter(jaeger.Options{
		CollectorEndpoint: "http://localhost:14268/api/traces",
		ServiceName:       "Gopher Feeder",
	})
	if err != nil {
		log.Fatal(err)
	}
	trace.RegisterExporter(je)
	// You should reduce sampler in a production app.
	trace.ApplyConfig(trace.Config{DefaultSampler: trace.AlwaysSample()})

	log.Println("Start demo server...")
	http.HandleFunc("/", foodHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
