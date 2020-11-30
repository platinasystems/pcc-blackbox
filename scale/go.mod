module github.com/platinasystems/scale

go 1.13

require (
	github.com/docker/docker v1.13.1
	github.com/platinasystems/go-common v0.0.0-20201125102358-64d82556ad0a
	github.com/platinasystems/pcc-blackbox v1.6.0-rc1.0.20201130040709-92ff89d97ddd
	github.com/platinasystems/tiles v1.3.1-rc1.0.20201127171947-0c4abb02d113
)

replace github.com/platinasystems/pcc-blackbox => ../../pcc-blackbox6
