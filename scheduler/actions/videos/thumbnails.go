/*
 * Copyright (c) 2018. Abstrium SAS <team (at) pydio.com>
 * This file is part of Pydio Cells.
 *
 * Pydio Cells is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * Pydio Cells is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Pydio Cells.  If not, see <http://www.gnu.org/licenses/>.
 *
 * The latest code can be found at <https://pydio.com>.
 */

package videos

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/image/colornames"
	"google.golang.org/protobuf/proto"

	"github.com/pydio/cells/v4/common"
	"github.com/pydio/cells/v4/common/client/grpc"
	"github.com/pydio/cells/v4/common/forms"
	"github.com/pydio/cells/v4/common/log"
	"github.com/pydio/cells/v4/common/nodes"
	"github.com/pydio/cells/v4/common/nodes/models"
	"github.com/pydio/cells/v4/common/proto/jobs"
	"github.com/pydio/cells/v4/common/proto/tree"
	json "github.com/pydio/cells/v4/common/utils/jsonx"
	"github.com/pydio/cells/v4/common/utils/uuid"
	"github.com/pydio/cells/v4/scheduler/actions"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

const (
	MetadataThumbnails      = "imageThumbnails"
	MetadatavideoDimensions = "imageDimensions"

	MetadataCompatIsvideo                 = "is_image"
	MetadataCompatvideoWidth              = "image_width"
	MetadataCompatvideoHeight             = "image_height"
	MetadataCompatvideoReadableDimensions = "readable_dimension"
	MetadataCompatImagePreview            = "ImagePreview"
)

var (
	thumbnailsActionName = "actions.videos.thumbnails"
)

type ThumbnailData struct {
	Format string `json:"format"`
	Size   int    `json:"size"`
	Id     string `json:"id"`
}

type ThumbnailsMeta struct {
	Processing bool
	Thumbnails []ThumbnailData `json:"thumbnails"`
}

type ThumbnailExtractor struct {
	common.RuntimeHolder
	thumbSizes map[string]int
	metaClient tree.NodeReceiverClient
}

// GetDescription returns action description
func (t *ThumbnailExtractor) GetDescription(_ ...string) actions.ActionDescription {
	return actions.ActionDescription{
		ID:                thumbnailsActionName,
		Label:             "Create Thumbs",
		Icon:              "video-filter",
		Description:       "Create thumbnails on video creation/modification",
		SummaryTemplate:   "",
		HasForm:           true,
		Category:          actions.ActionCategoryContents,
		InputDescription:  "Single-selection of file. Temporary and zero-bytes will be ignored",
		OutputDescription: "Input file with updated metadata",
	}
}

// GetParametersForm returns a UX form
func (t *ThumbnailExtractor) GetParametersForm() *forms.Form {
	return &forms.Form{Groups: []*forms.Group{
		{
			Fields: []forms.Field{
				&forms.FormField{
					Name:        "ThumbSizes",
					Type:        forms.ParamTextarea,
					Label:       "Sizes",
					Description: "A JSON map describing each thumbnail to be created",
					Default:     "",
					Mandatory:   false,
					Editable:    true,
				},
			},
		},
	}}
}

// GetName returns this action unique identifier.
func (t *ThumbnailExtractor) GetName() string {
	return thumbnailsActionName
}

// Init passes parameters to the action.
func (t *ThumbnailExtractor) Init(job *jobs.Job, action *jobs.Action) error {
	if action.Parameters != nil {
		t.thumbSizes = make(map[string]int)
		if params, ok := action.Parameters["ThumbSizes"]; ok {
			if e := json.Unmarshal([]byte(params), &t.thumbSizes); e != nil {
				for i, s := range strings.Split(params, ",") {
					parsed, _ := strconv.ParseInt(s, 10, 32)
					t.thumbSizes[fmt.Sprintf("%d", i)] = int(parsed)
				}
			}
		}
	} else {
		t.thumbSizes = map[string]int{"sm": 300}
	}
	if !nodes.IsUnitTestEnv {
		t.metaClient = tree.NewNodeReceiverClient(grpc.GetClientConnFromCtx(t.GetRuntimeContext(), common.ServiceMeta))
	}
	return nil
}

// Run the actual action code.
func (t *ThumbnailExtractor) Run(ctx context.Context, _ *actions.RunnableChannels, input jobs.ActionMessage) (jobs.ActionMessage, error) {

	if len(input.Nodes) == 0 || input.Nodes[0].Size == -1 || input.Nodes[0].Etag == common.NodeFlagEtagTemporary {
		// Nothing to do
		log.Logger(ctx).Debug("[THUMB EXTRACTOR] task ignored")
		return input.WithIgnore(), nil
	}

	log.Logger(ctx).Debug("[THUMB EXTRACTOR] Resizing video...")
	node := input.Nodes[0]
	err := t.resize(ctx, node, t.thumbSizes)
	if err != nil {
		return input.WithError(err), err
	}

	output := input
	output.Nodes[0] = node
	log.TasksLogger(ctx).Info("Created thumbnails for "+node.GetPath(), node.ZapPath())
	output.AppendOutput(&jobs.ActionOutput{Success: true})

	return output, nil
}

func displayMemStat(_ context.Context, _ string) {
	//var m runtime.MemStats
	//runtime.ReadMemStats(&m)
	//log.Logger(ctx).Info("---- MEMORY USAGE AT: "+position, zap.Uint64("Alloc", m.Alloc/1024/1024), zap.Uint64("TotalAlloc", m.TotalAlloc/1024/1024), zap.Uint64("Sys", m.Sys/1024/1024), zap.Uint32("NumGC", m.NumGC))
	//stdlog.Printf("%s : \nAlloc = %v\nTotalAlloc = %v\nSys = %v\nNumGC = %v\n\n", position, m.Alloc / 1024 / 1024, m.TotalAlloc / 1024 / 1024, m.Sys / 1024, m.NumGC)
}

func ExampleReadFrameAsJpeg(inFileName string, frameNum int) io.Reader {
	buf := bytes.NewBuffer(nil)
	err := ffmpeg.Input(inFileName).
		Filter("select", ffmpeg.Args{fmt.Sprintf("gte(n,%d)", frameNum)}).
		Output("pipe:", ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg"}).
		WithOutput(buf, os.Stdout).
		Run()
	if err != nil {
		panic(err)
	}
	return buf
}

func CreateDirIfNotExist(ctx context.Context, filePath string) {
	fileDir, fileName := filepath.Split(filePath)
	s, err := os.Stat(fileDir)
	if err != nil {
		os.Mkdir(fileDir, 0755)
		log.Logger(ctx).Debug(fmt.Sprintf("create parrent dir %s for path  %s:  %t", fileDir, fileName, s.IsDir()))
	}
}

func deleteFile(ctx context.Context, filePath string) {
	err := os.Remove(filePath)

	if err != nil {
		log.Logger(ctx).Debug("remove file error " + err.Error())
	}
}

func (t *ThumbnailExtractor) resize(ctx context.Context, node *tree.Node, sizes map[string]int) error {
	displayMemStat(ctx, "START RESIZE")
	// Open the test video.
	if !node.HasSource() {
		return fmt.Errorf("node does not have enough metadata for Resize (missing Source data)")
	}
	log.Logger(ctx).Debug("[THUMB EXTRACTOR] Getting object content", zap.String("Path", node.Path), zap.Int64("Size", node.Size))
	var reader io.ReadCloser
	var err error
	var errPath string
	var tempPath string
	if localPath := getNodeLocalPath(node); len(localPath) > 0 {
		reader, err = os.Open(localPath)
		errPath = localPath
		tempPath = localPath
		log.Logger(ctx).Debug("customize ", zap.String("localpath", localPath))
	} else {
		// Security in case Router is not transmitting nodes immutably
		routerNode := proto.Clone(node).(*tree.Node)
		reader, err = getRouter(t.GetRuntimeContext()).GetObject(ctx, routerNode, &models.GetRequestData{Length: -1})
		errPath = routerNode.Path
		tempPath = routerNode.Path
		log.Logger(ctx).Debug("customize ", zap.String("routePath", routerNode.Path))
	}

	if err != nil {
		return errors.Wrap(err, errPath)
	}
	log.Logger(ctx).Info("customize before")
	log.Logger(ctx).Info("customize" + tempPath)

	finfo, err := os.Stat(tempPath)
	if err != nil {
		// no such file or dir
		log.Logger(ctx).Debug("customize no such file")

	} else {
		log.Logger(ctx).Debug("customize file exists", zap.String("finfo", finfo.Name()))
	}

	defer reader.Close()

	displayMemStat(ctx, "BEFORE DECODE")

	buf := new(bytes.Buffer)
	buf.ReadFrom(reader)

	uuidFilePath := "/tmp/pydio/" + uuid.New()
	CreateDirIfNotExist(ctx, uuidFilePath)
	ioutil.WriteFile(uuidFilePath, buf.Bytes(), 0666)

	fileReader := ExampleReadFrameAsJpeg(uuidFilePath, 100)

	log.Logger(ctx).Info("customize after")
	thumb, err := imaging.Decode(fileReader)
	fileReader = nil
	buf = nil

	if err != nil {
		log.Logger(ctx).Info("customize ", zap.String("errInfo: ", err.Error()))
		return errors.Wrap(err, errPath)
	}

	displayMemStat(ctx, "AFTER DECODE")

	// Extract dimensions
	bounds := thumb.Bounds()
	width := bounds.Max.X
	height := bounds.Max.Y
	// Send update event right now
	node.MustSetMeta(MetadatavideoDimensions, struct {
		Width  int
		Height int
	}{
		Width:  width,
		Height: height,
	})
	node.MustSetMeta(MetadataCompatIsvideo, true)
	node.MustSetMeta(MetadataCompatImagePreview, true)
	node.MustSetMeta(MetadataThumbnails, &ThumbnailsMeta{Processing: true})
	node.MustSetMeta(MetadataCompatvideoHeight, height)
	node.MustSetMeta(MetadataCompatvideoWidth, width)
	node.MustSetMeta(MetadataCompatvideoReadableDimensions, fmt.Sprintf("%dpx X %dpx", width, height))

	if _, err = t.metaClient.UpdateNode(ctx, &tree.UpdateNodeRequest{From: node, To: node}); err != nil {
		return errors.Wrap(err, errPath)
	}

	log.Logger(ctx).Debug("Thumbnails - Extracted dimension and saved in metadata", zap.Any("dimension", bounds))
	meta := &ThumbnailsMeta{}

	for metaId, size := range sizes {

		if (metaId == "md" || metaId == "lg") && (width <= size && height <= size) {
			log.Logger(ctx).Debug("Ignoring thumbnails for size as original is smaller", zap.Any("dimension", bounds), zap.Any("thumbSize", size))
			continue
		}

		displayMemStat(ctx, "BEFORE WRITE SIZE FROM SRC")
		updateMeta, err := t.writeSizeFromSrc(ctx, thumb, node, size)
		if err != nil {
			// Remove processing state from Metadata
			node.MustSetMeta(MetadataThumbnails, nil)
			t.metaClient.UpdateNode(ctx, &tree.UpdateNodeRequest{From: node, To: node})
			return errors.Wrap(err, errPath)
		}
		displayMemStat(ctx, "AFTER WRITE SIZE FROM SRC")
		if updateMeta {
			meta.Thumbnails = append(meta.Thumbnails, ThumbnailData{
				Id:     metaId,
				Format: "jpg",
				Size:   size,
			})
		} else {
			log.Logger(ctx).Debug("customize ", zap.String("path", tempPath))
		}
	}

	thumb = nil

	runtime.GC()

	displayMemStat(ctx, "AFTER TRIGGERING GC")

	if (meta != &ThumbnailsMeta{}) {
		node.MustSetMeta(MetadataThumbnails, meta)
	} else {
		node.MustSetMeta(MetadataThumbnails, nil)
	}

	_, err = t.metaClient.UpdateNode(ctx, &tree.UpdateNodeRequest{From: node, To: node})
	if err != nil {
		err = errors.Wrap(err, errPath)
	}

	deleteFile(ctx, uuidFilePath)

	return err
}

func (t *ThumbnailExtractor) writeSizeFromSrc(ctx context.Context, img image.Image, node *tree.Node, targetSize int) (bool, error) {

	localTest := false
	localFolder := ""

	objectName := fmt.Sprintf("%s-%d.jpg", node.Uuid, targetSize)

	if localFolder = node.GetStringMeta(common.MetaNamespaceNodeTestLocalFolder); localFolder != "" {
		localTest = true
	}
	logger := log.Logger(ctx)

	if !localTest {

		dsi, e := getRouter(t.GetRuntimeContext()).GetClientsPool().GetDataSourceInfo(common.PydioThumbstoreNamespace)
		if e != nil || dsi.Client == nil {
			logger.Error("Cannot find ds info for thumbnail store", zap.Error(e))
			return false, e
		}
		// First Check if thumb already exists with same original etag
		oi, check := dsi.Client.StatObject(ctx, dsi.ObjectsBucket, objectName, nil)
		logger.Debug("Object Info", zap.Any("object", oi), zap.Error(check))
		if check == nil {
			foundOriginal := oi.Metadata.Get("X-Amz-Meta-Original-Etag")
			if len(foundOriginal) > 0 && foundOriginal == node.Etag {
				// No update necessary
				logger.Debug("Ignoring Resize: thumb already exists in store", zap.Any("original", oi))
				return true, nil
			}
		}

	}

	logger.Debug("WriteSizeFromSrc", zap.String("nodeUuid", node.Uuid))
	var dst *image.NRGBA
	if img.Bounds().Max.X >= img.Bounds().Max.Y {
		// Resize the cropped video to width = 256px preserving the aspect ratio.
		dst = imaging.Resize(img, targetSize, 0, imaging.Lanczos)
	} else {
		// Resize the cropped video to height = 256px preserving the aspect ratio.
		dst = imaging.Resize(img, 0, targetSize, imaging.Lanczos)
	}
	ol := imaging.New(dst.Bounds().Dx(), dst.Bounds().Dy(), colornames.Lightgrey)
	ol = imaging.Overlay(ol, dst, image.Pt(0, 0), 1.0)
	dst = nil
	runtime.GC()

	displayMemStat(ctx, "BEFORE ENCODE")
	var thumbBytes []byte
	buf := bytes.NewBuffer(thumbBytes)
	err := imaging.Encode(buf, ol, imaging.JPEG)
	ol = nil
	runtime.GC()
	if err != nil {
		return false, err
	}

	displayMemStat(ctx, "AFTER ENCODE")

	if !localTest {

		requestMeta := map[string]string{
			common.XContentType:        "video/jpeg",
			"X-Amz-Meta-Original-Etag": node.Etag,
		}
		logger.Debug("Writing thumbnail to thumbs bucket", zap.Any("video size", targetSize))
		displayMemStat(ctx, "BEFORE PUT OBJECT WITH CONTEXT")
		tCtx, tNode, e := getThumbLocation(t.GetRuntimeContext(), ctx, objectName)
		if e != nil {
			return false, e
		}
		tNode.Size = int64(buf.Len())
		oi, err := getRouter(t.GetRuntimeContext()).PutObject(tCtx, tNode, buf, &models.PutRequestData{
			Size:     tNode.Size,
			Metadata: requestMeta,
		})
		if err != nil {
			return false, err
		} else {
			logger.Debug("Finished putting thumb for size", zap.Int64("written", oi.Size), zap.Int("size ", targetSize))
		}
		displayMemStat(ctx, "AFTER PUT OBJECT WITH CONTEXT")

	} else {
		e := os.WriteFile(filepath.Join(localFolder, objectName), buf.Bytes(), 0755)
		if e != nil {
			return false, e
		}
	}

	logger.Debug("WriteSizeFromSrc: END", zap.String("nodeUuid", node.Uuid))

	return true, err

}

func getNodeLocalPath(node *tree.Node) string {
	if localFolder := node.GetStringMeta(common.MetaNamespaceNodeTestLocalFolder); localFolder != "" {
		baseName := node.GetStringMeta(common.MetaNamespaceNodeName)
		targetFileName := filepath.Join(localFolder, baseName)
		return targetFileName
	}
	return ""
}
