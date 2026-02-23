import express from "express";
import { Sandbox } from "e2b";

const app = express();
app.use(express.json());

app.post("/run", async (req, res) => {
  const { code, language, testCases } = req.body;

  if (code == null || typeof code !== "string") {
    return res.status(400).json({ error: "code is required and must be a string" });
  }
  if (!Array.isArray(testCases)) {
    return res.status(400).json({ error: "testCases is required and must be an array" });
  }

  const lang = (language || "python").toLowerCase();
  const langConfig = {
    python: { filename: "solution.py", runCmd: "python3 solution.py" },
    go: { filename: "main.go", compileCmd: "go build -o main main.go", runCmd: "./main" },
    javascript: { filename: "solution.js", runCmd: "node solution.js" },
    js: { filename: "solution.js", runCmd: "node solution.js" },
    node: { filename: "solution.js", runCmd: "node solution.js" },
    java: { filename: "Main.java", runCmd: "java Main", compileCmd: "javac Main.java" },
  };

  const config = langConfig[lang];
  if (!config) {
    return res.status(400).json({ error: "Unsupported language. Use: python, go, javascript, js, java" });
  }

  const { filename, runCmd, compileCmd } = config;
  let sandbox;

  try {
    sandbox = await Sandbox.create("langsupport-dev");

    await sandbox.files.write(filename, code);

    if (compileCmd) {
      const compile = await sandbox.commands.run(compileCmd, {
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

    // Write all inputs to unique files (enables parallel execution)
    await Promise.all(
      testCases.map((tc, i) => sandbox.files.write(`input_${i}.txt`, tc.input ?? ""))
    );

    // Run all test cases in parallel (with logging to verify parallel execution)
    const runStart = Date.now();
    console.log(`[parallel] Starting ${testCases.length} test(s) at ${new Date().toISOString()}`);

    const runPromises = testCases.map((_, i) =>
      sandbox.commands
        .run(`${runCmd} < input_${i}.txt`)
        .then((result) => {
          console.log(`[parallel] Test ${i} completed in ${Date.now() - runStart}ms`);
          return result;
        })
    );
    const runResults = await Promise.all(runPromises);

    const totalMs = Date.now() - runStart;
    console.log(`[parallel] All ${testCases.length} test(s) done in ${totalMs}ms total`);

    const results = runResults.map((result, i) => {
      const tc = testCases[i];
      if (result.exitCode !== 0) {
        return {
          input: tc.input,
          expected: tc.expected,
          output: result.stderr.trim(),
          passed: false,
        };
      }
      const output = result.stdout.trim();
      const passed = output === tc.expected.trim();
      return {
        input: tc.input,
        expected: tc.expected,
        output,
        passed,
      };
    });

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
