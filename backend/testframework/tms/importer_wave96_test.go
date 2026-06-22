package tms_test

// Тесты модуля TMS (волна 96): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

func buildSampleExcel(t *testing.T) *bytes.Buffer {
	t.Helper()
	f := excelize.NewFile()
	defer f.Close()
	sheet := f.GetSheetName(0)
	require.NoError(t, f.SetCellValue(sheet, "A1", "Title"))
	require.NoError(t, f.SetCellValue(sheet, "B1", "Priority"))
	require.NoError(t, f.SetCellValue(sheet, "C1", "Type"))
	require.NoError(t, f.SetCellValue(sheet, "A2", "Excel import case"))
	require.NoError(t, f.SetCellValue(sheet, "B2", "high"))
	require.NoError(t, f.SetCellValue(sheet, "C2", "manual"))
	var buf bytes.Buffer
	require.NoError(t, f.Write(&buf))
	return &buf
}

// TestImportFromExcel проверяет сценарий: импорт from excel.
func TestImportFromExcel(t *testing.T) {
	buf := buildSampleExcel(t)
	inputs, err := tms.ImportFromExcel(buf, "", nil)
	require.NoError(t, err)
	require.Len(t, inputs, 1)
	assert.Equal(t, "Excel import case", inputs[0].Title)
	assert.Equal(t, "high", inputs[0].Priority)
}

// TestImportFromExcel_EmptySheet проверяет сценарий: импорт from excel пустой sheet.
func TestImportFromExcel_EmptySheet(t *testing.T) {
	f := excelize.NewFile()
	defer f.Close()
	var buf bytes.Buffer
	require.NoError(t, f.Write(&buf))
	_, err := tms.ImportFromExcel(&buf, "", nil)
	require.Error(t, err)
}

// TestImportFromExcel_InvalidData проверяет отклонение при импорт from excel некорректный data.
func TestImportFromExcel_InvalidData(t *testing.T) {
	_, err := tms.ImportFromExcel(bytes.NewReader([]byte("not-xlsx")), "", nil)
	require.Error(t, err)
}
