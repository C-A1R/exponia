package logger

import (
	"fmt"
	"io"
	"strings"
)

const bannerWidth = 60

func WriteStartupBanner(
	output io.Writer,
	version string,
	commit string,
	buildTime string,
) error {
	separator := strings.Repeat("=", bannerWidth)
	titlePadding := strings.Repeat(" ", (bannerWidth-len("EXPONIA"))/2)

	var banner strings.Builder
	banner.WriteString(separator)
	banner.WriteByte('\n')
	banner.WriteString(titlePadding)
	banner.WriteString("EXPONIA\n")
	banner.WriteString(separator)
	banner.WriteByte('\n')
	fmt.Fprintf(&banner, "version:    %s\n", version)
	fmt.Fprintf(&banner, "commit:     %s\n", commit)
	fmt.Fprintf(&banner, "build_time: %s\n", buildTime)
	banner.WriteString(separator)
	banner.WriteByte('\n')

	_, err := io.WriteString(output, banner.String())
	return err
}
