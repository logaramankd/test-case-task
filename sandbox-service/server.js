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


    if (lang === "python") {
      await sandbox.commands.run(`cat << 'EOF' > solution.py
${code}
EOF`);
    }

    if (lang === "go") {
      await sandbox.commands.run(`cat << 'EOF' > main.go
${code}
EOF`);
    }

    if (lang === "javascript" || lang === "js" || lang === "node") {
      await sandbox.commands.run(`cat << 'EOF' > solution.js
${code}
EOF`);
    }

    if (lang === "java") {
      await sandbox.commands.run(`cat << 'EOF' > Main.java
${code}
EOF`);

      const compile = await sandbox.commands.run(`javac Main.java`);

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
      await sandbox.close();
    }
  }
});

app.listen(3001, () => {
  console.log("Sandbox service running on port 3001");
});
