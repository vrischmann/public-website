// Code generated by templ@(devel) DO NOT EDIT.

package templates

//lint:file-ignore SA4006 This context is only used if a nested component is present.

import "github.com/a-h/templ"
import "context"
import "io"
import "bytes"

func Resume(skills templ.Component, experience []templ.Component, sideProjects templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_1 := templ.GetChildren(ctx)
		if var_1 == nil {
			var_1 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, err = templBuffer.WriteString("<div class=\"resume\"><div class=\"resume-header\"><div class=\"title\"><h1>")
		if err != nil {
			return err
		}
		var_2 := `Vincent Rischmann`
		_, err = templBuffer.WriteString(var_2)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</h1><h2>")
		if err != nil {
			return err
		}
		var_3 := `Staff engineer`
		_, err = templBuffer.WriteString(var_3)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</h2></div><div class=\"links\"><a href=\"mailto:vincent@rischmann.fr\" class=\"envelope\">")
		if err != nil {
			return err
		}
		var_4 := `vincent@rischmann.fr`
		_, err = templBuffer.WriteString(var_4)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</a><i class=\"fa-solid fa-envelope\"></i><a href=\"https://rischmann.fr\">")
		if err != nil {
			return err
		}
		var_5 := `rischmann.fr`
		_, err = templBuffer.WriteString(var_5)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</a><i class=\"fa-solid fa-globe\"></i><a href=\"https://github.com/vrischmann\">")
		if err != nil {
			return err
		}
		var_6 := `GitHub`
		_, err = templBuffer.WriteString(var_6)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</a><i class=\"fa-brands fa-github\"></i><a href=\"/assets/resume.pdf\">")
		if err != nil {
			return err
		}
		var_7 := `PDF`
		_, err = templBuffer.WriteString(var_7)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</a><i class=\"fa-solid fa-file\"></i></div></div><div class=\"resume-summary\"><h2>")
		if err != nil {
			return err
		}
		var_8 := `Summary`
		_, err = templBuffer.WriteString(var_8)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</h2><p>")
		if err != nil {
			return err
		}
		var_9 := `I am a Staff engineer with 10+ years of experience building distributed systems, high-throughput webservices and data processing pipelines.`
		_, err = templBuffer.WriteString(var_9)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</p></div><div class=\"resume-skills\">")
		if err != nil {
			return err
		}
		err = skills.Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</div><div class=\"resume-experience\"><h2>")
		if err != nil {
			return err
		}
		var_10 := `Work experience`
		_, err = templBuffer.WriteString(var_10)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</h2>")
		if err != nil {
			return err
		}
		for _, workExperience := range experience {
			_, err = templBuffer.WriteString("<div class=\"work-experience\">")
			if err != nil {
				return err
			}
			err = workExperience.Render(ctx, templBuffer)
			if err != nil {
				return err
			}
			_, err = templBuffer.WriteString("</div>")
			if err != nil {
				return err
			}
		}
		_, err = templBuffer.WriteString("</div><div class=\"resume-side-projects\">")
		if err != nil {
			return err
		}
		err = sideProjects.Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</div><div class=\"resume-interests\"><h2>")
		if err != nil {
			return err
		}
		var_11 := `Interests`
		_, err = templBuffer.WriteString(var_11)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</h2><p>")
		if err != nil {
			return err
		}
		var_12 := `Movies, TV shows, listening to music, podcasts and audiobooks.`
		_, err = templBuffer.WriteString(var_12)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</p><p>")
		if err != nil {
			return err
		}
		var_13 := `Video games, programming, discovering new things.`
		_, err = templBuffer.WriteString(var_13)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</p></div><div class=\"resume-mobile-links\"><h2>")
		if err != nil {
			return err
		}
		var_14 := `Contacts`
		_, err = templBuffer.WriteString(var_14)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</h2><ul class=\"links\"><li><a href=\"mailto:vincent@rischmann.fr\" class=\"envelope\">")
		if err != nil {
			return err
		}
		var_15 := `vincent@rischmann.fr`
		_, err = templBuffer.WriteString(var_15)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</a></li><li><a href=\"https://rischmann.fr\">")
		if err != nil {
			return err
		}
		var_16 := `rischmann.fr`
		_, err = templBuffer.WriteString(var_16)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</a></li><li><a href=\"https://github.com/vrischmann\">")
		if err != nil {
			return err
		}
		var_17 := `GitHub`
		_, err = templBuffer.WriteString(var_17)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</a></li><li><a href=\"https://rischmann.fr/resume.pdf\">")
		if err != nil {
			return err
		}
		var_18 := `PDF`
		_, err = templBuffer.WriteString(var_18)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</a></li></ul></div></div>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = io.Copy(w, templBuffer)
		}
		return err
	})
}

func ResumePage(title string, assets Assets, body templ.Component) templ.Component {
	return templ.ComponentFunc(func(ctx context.Context, w io.Writer) (err error) {
		templBuffer, templIsBuffer := w.(*bytes.Buffer)
		if !templIsBuffer {
			templBuffer = templ.GetBuffer()
			defer templ.ReleaseBuffer(templBuffer)
		}
		ctx = templ.InitializeContext(ctx)
		var_19 := templ.GetChildren(ctx)
		if var_19 == nil {
			var_19 = templ.NopComponent
		}
		ctx = templ.ClearChildren(ctx)
		_, err = templBuffer.WriteString("<html>")
		if err != nil {
			return err
		}
		err = headerComponent(title, assets).Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("<script src=\"https://kit.fontawesome.com/bb474c1b63.js\" crossorigin=\"anonymous\">")
		if err != nil {
			return err
		}
		var_20 := ``
		_, err = templBuffer.WriteString(var_20)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</script>")
		if err != nil {
			return err
		}
		err = body.Render(ctx, templBuffer)
		if err != nil {
			return err
		}
		_, err = templBuffer.WriteString("</html>")
		if err != nil {
			return err
		}
		if !templIsBuffer {
			_, err = io.Copy(w, templBuffer)
		}
		return err
	})
}
