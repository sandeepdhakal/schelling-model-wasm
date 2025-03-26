function data_for_plotly(data) {
  res = [];
  data.forEach(function (iter) {
    xs = [];
    ys = [];
    ts = [];
    iter.forEach(function (a, _) {
      xs.push(a[0]);
      ys.push(a[1]);
      ts.push(a[2]);
    });
    res.push({ x: xs, y: ys, t: ts });
  });
  return res;
}

function data_for_observable(res) {
  data = [];
  res.forEach(function (iter, idx) {
    iter.forEach(function (a, _) {
      data.push({
        x: a[0],
        y: a[1],
        type: a[2],
        iteration: idx,
      });
    });
  });
  return data;
}

function observable_plot(data) {
  dim = 250;

  const plot = Plot.plot(
    (() => {
      const n = 3; // number of columns
      const keys = Array.from(d3.union(data.map((d) => d.iteration)));
      const index = new Map(keys.map((key, i) => [key, i]));
      const fx = (key) => index.get(key) % n;
      const fy = (key) => Math.floor(index.get(key) / n);

      return {
        height: (dim * keys.length) / n,
        width: dim * n,
        axis: null,
        grid: false,
        color: { type: "categorical" },
        inset: 5,
        marginTop: 15,
        fx: { padding: 0.1 },
        fy: { padding: 0.15 },
        marks: [
          Plot.dot(data, {
            x: "x",
            y: "y",
            stroke: "type",
            symbol: "type",
            fx: (d) => fx(d.iteration),
            fy: (d) => fy(d.iteration),
            r: 3,
          }),
          Plot.text(keys, {
            fx,
            fy,
            frameAnchor: "top-left",
            dx: 5,
            dy: -14,
            text: (d) => `iteration ${d}`,
            fontSize: 14,
          }),
          Plot.frame(),
        ],
      };
    })(),
  );
  return plot;
}

function observable_facet_plot(res) {
  data = data_for_observable(res);
  plot = observable_plot(data);
  const div = document.querySelector("#myplot");
  div.innerHTML = "";
  div.append(plot);
}

function plotly_animated(res) {
  data = data_for_plotly(res);
  ids = Array.from({ length: data[0].x.length }, (x, i) => `${i}`);

  traces = [
    {
      x: data[0].x.slice(),
      y: data[0].y.slice(),
      ids: ids.slice(),
      mode: "markers",
      marker: {
        color: data[0].t.map((x) => (x ? "green" : "orange")).slice(),
      },
    },
  ];

  frames_ = [];
  for (i = 0; i < data.length; i++) {
    frames_.push({
      name: i,
      data: [
        {
          ids: ids,
          x: data[i].x,
          y: data[i].y,
          marker: {
            color: data[i].t.map((x) => (x ? "green" : "orange")),
          },
        },
      ],
    });
  }

  var frame_duration = 1000;
  var sliderSteps = [];
  for (i = 0; i < data.length; i++) {
    sliderSteps.push({
      method: "animate",
      label: i,
      args: [
        [i],
        {
          mode: "immediate",
          transition: { duration: frame_duration },
          frame: { duration: frame_duration, redraw: false },
        },
      ],
    });
  }

  axes = {
    visible: false,
  };

  var layout = {
    width: 800,
    height: 800,
    xaxis: axes,
    yaxis: axes,
    hovermode: "closest",
    updatemenus: [
      {
        x: 0,
        y: 0,
        yanchor: "top",
        xanchor: "left",
        showactive: false,
        direction: "left",
        type: "buttons",
        pad: { t: 87, r: 10 },
        buttons: [
          {
            method: "animate",
            args: [
              null,
              {
                mode: "immediate",
                fromcurrent: true,
                transition: { duration: frame_duration },
                frame: { duration: frame_duration, redraw: false },
              },
            ],
            label: "Play",
          },
          {
            method: "animate",
            args: [
              [null],
              {
                mode: "immediate",
                transition: { duration: 0 },
                frame: { duration: 0, redraw: false },
              },
            ],
            label: "Pause",
          },
        ],
      },
    ],
    // Finally, add the slider and use `pad` to position it
    // nicely next to the buttons.
    sliders: [
      {
        pad: { l: 130, t: 55 },
        currentvalue: {
          visible: true,
          prefix: "Iteration:",
          xanchor: "right",
          font: { size: 20, color: "#666" },
        },
        steps: sliderSteps,
      },
    ],
  };

  const div = document.querySelector("#myplot");
  div.innerHTML = "";
  Plotly.newPlot("myplot", {
    data: traces,
    layout: layout,
    frames: frames_,
    config: { responsive: true },
  });
}

function sim_and_vis(plot_type) {
  na = 1000;
  nn = 10;
  ns = parseInt(document.getElementById("happy_ratio").value);
  nr = 10;

  // document.getElementById("observable").disabled = true;
  document.getElementById("plotly").disabled = true;
  document.getElementById("loader").style.display = "inline-block";

  const worker = new Worker("js/wasm_worker.js");
  worker.postMessage([na, nn, ns, nr]);
  worker.onmessage = (e) => {
    if (plot_type == "observable") {
      observable_facet_plot(e.data);
    } else {
      plotly_animated(e.data);
    }

    // document.getElementById("observable").disabled = false;
    document.getElementById("plotly").disabled = false;
    document.getElementById("loader").style.display = "none";
  };
}

// function sim_with_observable() {
//   sim_and_vis("observable");
// }

function sim_with_plotly() {
  sim_and_vis("plotly");
}
