package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"jobsearcher/internal/config"
	"jobsearcher/internal/domain"
	"jobsearcher/internal/filter"
	"jobsearcher/internal/matching"
	"jobsearcher/internal/sources"
)

func webJobEligible(j domain.Job, minimumFit int) bool {
	return j.WorkplaceType == "remote" && j.LocationEligible == domain.LocationEligible && j.FitScore >= minimumFit && j.SeniorityMatch > 20
}
