export default function ElementSelector({ label, value, onChange, options }) {
  return (
    <div className="mb-4">
      <label className="block mb-1 text-gray-700 font-semibold">{label}</label>
      <select
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full p-2 border border-gray-300 rounded"
      >
        <option value="">-- Pilih elemen --</option>
        {options.map((el) => (
          <option key={el} value={el}>
            {el}
          </option>
        ))}
      </select>
    </div>
  );
}
