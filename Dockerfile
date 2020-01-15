FROM scratch

ADD cm2metric /cm2metric

ENTRYPOINT ["./cm2metric"]

