import express from "express";
import { Sandbox } from "e2b";

const app = express();
app.use(express.json());

app.post("/run", async (req, res) => {
  const { code, language, testCases } = req.body;
  const lang = (language || "python").toLowerCase();

  let sandbox;

  try {
    sandbox = await Sandbox.create("langsupport-dev");

    const results = [];


    let filename = "";
    let compileCommand = null;

    if (lang === "python") {
      filename = "solution.py";
    }

    if (lang === "go") {
      filename = "main.go";
    }

    if (lang === "javascript" || lang === "js" || lang === "node") {
      filename = "solution.js";
    }

    if (lang === "java") {
      filename = "Main.java";
      compileCommand = "javac Main.java";
    }

    if (!filename) {
      return res.status(400).json({ error: "Unsupported language" });
    }

    await sandbox.files.write(filename, code);

    if (compileCommand) {
      const compile = await sandbox.commands.run(compileCommand, {
        timeout: 5000,
      });

      if (compile.exitCode !== 0) {
        return res.json(
          testCases.map((tc) => ({
            input: tc.input,
            expected: tc.expected,
            output: compile.stderr.trim(),
            passed: false,
          }))
        );
      }
    }

    for (const tc of testCases) {
      let command = "";

      if (lang === "python") {
        command = `echo "${tc.input}" | python3 solution.py`;
      }

      if (lang === "go") {
        command = `echo "${tc.input}" | go run main.go`;
      }

      if (lang === "javascript" || lang === "js" || lang === "node") {
        command = `echo "${tc.input}" | node solution.js`;
      }

      if (lang === "java") {
        command = `echo "${tc.input}" | java Main`;
      }

      const result = await sandbox.commands.run(command);

      if (result.exitCode !== 0) {
        results.push({
          input: tc.input,
          expected: tc.expected,
          output: result.stderr.trim(),
          passed: false,
        });
        continue;
      }

      const output = result.stdout.trim();
      const passed = output === tc.expected.trim();

      results.push({
        input: tc.input,
        expected: tc.expected,
        output,
        passed,
      });
    }

    res.json(results);
  } catch (err) {
    res.status(500).json({ error: err.message });
  } finally {
    if (sandbox) {
      await sandbox.kill();
    }
  }
});

app.listen(3001, () => {
  console.log("Sandbox service running on port 3001");
});
