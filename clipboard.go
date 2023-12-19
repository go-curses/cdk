package cdk

const TypeClipboard CTypeTag = "cdk-clipboard"

func init() {
	_ = TypesManager.AddType(TypeClipboard, func() interface{} { return nil })
}

// Clipboard Hierarchy:
//
//	Object
//	  +- Clipboard
type Clipboard interface {
	Object

	GetText() (text string)
	SetText(text string)
	Copy(text string)
	Paste(text string)
}

var _ Clipboard = (*CClipboard)(nil)

type CClipboard struct {
	CObject

	screen Screen
}

func newClipboard(screen Screen) (clipboard *CClipboard) {
	clipboard = new(CClipboard)
	clipboard.screen = screen
	clipboard.Init()
	return
}

func (c *CClipboard) Init() (already bool) {
	if c.InitTypeItem(TypeClipboard, c) {
		return true
	}
	c.CObject.Init()
	_ = c.InstallProperty(PropertyText, StringProperty, true, "")
	return false
}

// SetText updates the clipboard's cache of pasted content
func (c *CClipboard) SetText(text string) {
	if err := c.SetStringProperty(PropertyText, text); err != nil {
		c.LogErr(err)
	}
}

// GetText retrieves the clipboard's cache of pasted content
func (c *CClipboard) GetText() (text string) {
	if c.screen.HostClipboardEnabled() {
		if v, ok := c.screen.PasteFromClipboard(); ok {
			c.LogDebug("updated from host clipboard value: \"%v\"", v)
			c.SetText(v)
			text = v
			return
		}
	}
	var err error
	if text, err = c.GetStringProperty(PropertyText); err != nil {
		c.LogErr(err)
	}
	return
}

// Copy updates the clipboard's cache of pasted content and passes the copy
// event to the underlying operating system (if supported) using OSC52 terminal
// sequences
func (c *CClipboard) Copy(text string) {
	c.SetText(text)
	c.Emit(SignalCopy, c, text)
	c.LogDebug("text: \"%v\"", text)
	c.screen.CopyToClipboard(text)
}

// Paste updates the clipboard's cache of pasted content and emits a "Paste"
// event itself
func (c *CClipboard) Paste(text string) {
	c.SetText(text)
	c.Emit(SignalPaste, c, text)
	c.LogDebug("text: \"%v\"", text)
}

const SignalCopy Signal = "copy"

const SignalPaste Signal = "paste"

const PropertyText Property = "text"
