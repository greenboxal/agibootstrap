export class ManifestBuilder {
    name = "";
    files = [];
    schema = {
        definitions: {},
    };
    withName(name) {
        this.name = name;
        return this;
    }
    withFile(name, hash) {
        this.files.push({ name, hash });
        return this;
    }
    withSchemaDefinition(schema) {
        if (!schema.$id) {
            throw new Error("Schema must have an $id");
        }
        schema = { ...schema };
        const defs = Object.keys(schema.definitions || {});
        if (defs.length > 0) {
            this.importSchemaDefinitions(schema, defs);
            schema.definitions = undefined;
        }
        this.schema.definitions[schema.$id] = schema;
        return this;
    }
    importSchemaDefinitions(from, names) {
        for (const name of names) {
            if (this.schema.definitions[name]) {
                continue;
            }
            const def = from.definitions[name];
            if (!def) {
                throw new Error(`Schema definition ${name} not found`);
            }
            this.withSchemaDefinition(def);
            if (def.$id != name) {
                this.schema.definitions[name] = { $ref: "#/definitions/" + def.$id };
            }
        }
    }
    build() {
        return {
            name: this.name,
            files: this.files,
            schema: this.schema,
        };
    }
}
//# sourceMappingURL=data:application/json;base64,eyJ2ZXJzaW9uIjozLCJmaWxlIjoiaW5kZXguanMiLCJzb3VyY2VSb290IjoiIiwic291cmNlcyI6WyIuLi8uLi9zcmMvaW5kZXgudHMiXSwibmFtZXMiOltdLCJtYXBwaW5ncyI6IkFBYUEsTUFBTSxPQUFPLGVBQWU7SUFDeEIsSUFBSSxHQUFXLEVBQUUsQ0FBQztJQUNsQixLQUFLLEdBQXlCLEVBQUUsQ0FBQztJQUNqQyxNQUFNLEdBQWU7UUFDakIsV0FBVyxFQUFFLEVBQUU7S0FDbEIsQ0FBQTtJQUVELFFBQVEsQ0FBQyxJQUFZO1FBQ2pCLElBQUksQ0FBQyxJQUFJLEdBQUcsSUFBSSxDQUFDO1FBQ2pCLE9BQU8sSUFBSSxDQUFDO0lBQ2hCLENBQUM7SUFFRCxRQUFRLENBQUMsSUFBWSxFQUFFLElBQVk7UUFDL0IsSUFBSSxDQUFDLEtBQUssQ0FBQyxJQUFJLENBQUMsRUFBQyxJQUFJLEVBQUUsSUFBSSxFQUFDLENBQUMsQ0FBQztRQUM5QixPQUFPLElBQUksQ0FBQztJQUNoQixDQUFDO0lBRUQsb0JBQW9CLENBQUMsTUFBa0I7UUFDbkMsSUFBSSxDQUFDLE1BQU0sQ0FBQyxHQUFHLEVBQUU7WUFDYixNQUFNLElBQUksS0FBSyxDQUFDLHlCQUF5QixDQUFDLENBQUM7U0FDOUM7UUFFRCxNQUFNLEdBQUcsRUFBQyxHQUFHLE1BQU0sRUFBQyxDQUFBO1FBRXBCLE1BQU0sSUFBSSxHQUFHLE1BQU0sQ0FBQyxJQUFJLENBQUMsTUFBTSxDQUFDLFdBQVcsSUFBSSxFQUFFLENBQUMsQ0FBQTtRQUVsRCxJQUFJLElBQUksQ0FBQyxNQUFNLEdBQUcsQ0FBQyxFQUFFO1lBQ2pCLElBQUksQ0FBQyx1QkFBdUIsQ0FBQyxNQUFNLEVBQUUsSUFBSSxDQUFDLENBQUE7WUFFMUMsTUFBTSxDQUFDLFdBQVcsR0FBRyxTQUFTLENBQUE7U0FDakM7UUFFRCxJQUFJLENBQUMsTUFBTSxDQUFDLFdBQVcsQ0FBQyxNQUFNLENBQUMsR0FBRyxDQUFDLEdBQUcsTUFBTSxDQUFDO1FBRTdDLE9BQU8sSUFBSSxDQUFDO0lBQ2hCLENBQUM7SUFFTyx1QkFBdUIsQ0FBQyxJQUFnQixFQUFFLEtBQWU7UUFDN0QsS0FBSyxNQUFNLElBQUksSUFBSSxLQUFLLEVBQUU7WUFDdEIsSUFBSSxJQUFJLENBQUMsTUFBTSxDQUFDLFdBQVcsQ0FBQyxJQUFJLENBQUMsRUFBRTtnQkFDL0IsU0FBUzthQUNaO1lBRUQsTUFBTSxHQUFHLEdBQUcsSUFBSSxDQUFDLFdBQVcsQ0FBQyxJQUFJLENBQUMsQ0FBQTtZQUVsQyxJQUFJLENBQUMsR0FBRyxFQUFFO2dCQUNOLE1BQU0sSUFBSSxLQUFLLENBQUMscUJBQXFCLElBQUksWUFBWSxDQUFDLENBQUE7YUFDekQ7WUFFRCxJQUFJLENBQUMsb0JBQW9CLENBQUMsR0FBRyxDQUFDLENBQUE7WUFFOUIsSUFBSSxHQUFHLENBQUMsR0FBRyxJQUFJLElBQUksRUFBRTtnQkFDakIsSUFBSSxDQUFDLE1BQU0sQ0FBQyxXQUFXLENBQUMsSUFBSSxDQUFDLEdBQUcsRUFBRSxJQUFJLEVBQUUsZ0JBQWdCLEdBQUcsR0FBRyxDQUFDLEdBQUcsRUFBRSxDQUFBO2FBQ3ZFO1NBQ0o7SUFDTCxDQUFDO0lBRUQsS0FBSztRQUNELE9BQU87WUFDSCxJQUFJLEVBQUUsSUFBSSxDQUFDLElBQUk7WUFDZixLQUFLLEVBQUUsSUFBSSxDQUFDLEtBQUs7WUFDakIsTUFBTSxFQUFFLElBQUksQ0FBQyxNQUFNO1NBQ3RCLENBQUE7SUFDTCxDQUFDO0NBQ0oifQ==